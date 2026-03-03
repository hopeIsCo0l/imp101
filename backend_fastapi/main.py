import json
import os
import re
import time
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Optional

from fastapi import Depends, FastAPI, File, Form, HTTPException, Request, UploadFile
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from jose import JWTError, jwt
from passlib.context import CryptContext
from pydantic import BaseModel, ConfigDict, EmailStr
from sqlalchemy import Boolean, DateTime, Float, Integer, String, Text, create_engine
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Mapped, Session, declarative_base, mapped_column, sessionmaker


Base = declarative_base()

ROLE_CANDIDATE = "candidate"
ROLE_RECRUITER = "recruiter"
ROLE_ADMIN = "administrator"
ROLE_SUPER_ADMIN = "super_admin"

STATUS_ACTIVE = "active"
STATUS_INACTIVE = "inactive"
STATUS_LOCKED = "locked"

JOB_DRAFT = "draft"
JOB_PUBLISHED = "published"
JOB_CLOSED = "closed"
JOB_ARCHIVED = "archived"

APP_SUBMITTED = "submitted"
APP_PARSED = "parsed"
APP_RANKED = "ranked"
APP_REJECTED = "rejected"

JWT_SECRET = os.getenv("JWT_SECRET", "development-secret-change-in-production")
JWT_ALG = "HS256"
JWT_TTL_HOURS = 24

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


def db_url() -> str:
    host = os.getenv("DB_HOST", "localhost")
    port = os.getenv("DB_PORT", "5432")
    user = os.getenv("DB_USER", "postgres")
    password = os.getenv("DB_PASSWORD", "postgres")
    name = os.getenv("DB_NAME", "imp101")
    return f"postgresql+psycopg2://{user}:{password}@{host}:{port}/{name}"


engine = create_engine(db_url(), future=True, pool_pre_ping=True)
SessionLocal = sessionmaker(bind=engine, autoflush=False, autocommit=False, future=True)


class User(Base):
    __tablename__ = "users"
    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True, nullable=False)
    password: Mapped[str] = mapped_column(Text, nullable=False)
    full_name: Mapped[str] = mapped_column(String(255), nullable=False, default="")
    phone: Mapped[str] = mapped_column(String(30), unique=True, nullable=False, default="")
    nationality: Mapped[str] = mapped_column(String(100), nullable=False, default="")
    date_of_birth: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=False), nullable=True)
    role: Mapped[str] = mapped_column(String(30), nullable=False, default=ROLE_CANDIDATE, index=True)
    status: Mapped[str] = mapped_column(String(30), nullable=False, default=STATUS_ACTIVE, index=True)
    is_email_verified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    is_phone_verified: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    failed_login_attempts: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    locked_until: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=False), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow, onupdate=datetime.utcnow)
    deleted_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=False), nullable=True)


class Job(Base):
    __tablename__ = "jobs"
    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    title: Mapped[str] = mapped_column(String(150), nullable=False, index=True)
    description: Mapped[str] = mapped_column(Text, nullable=False)
    required_skills: Mapped[str] = mapped_column(Text, nullable=False, default="")
    qualifications: Mapped[str] = mapped_column(Text, nullable=False, default="")
    criteria_weights: Mapped[str] = mapped_column(Text, nullable=False, default="")
    deadline: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=False), nullable=True, index=True)
    status: Mapped[str] = mapped_column(String(30), nullable=False, default=JOB_DRAFT, index=True)
    created_by: Mapped[int] = mapped_column(Integer, nullable=False, index=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow, onupdate=datetime.utcnow)


class Application(Base):
    __tablename__ = "applications"
    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    candidate_id: Mapped[int] = mapped_column(Integer, nullable=False, index=True)
    job_id: Mapped[int] = mapped_column(Integer, nullable=False, index=True)
    status: Mapped[str] = mapped_column(String(30), nullable=False, default=APP_SUBMITTED, index=True)
    cv_file_path: Mapped[str] = mapped_column(Text, nullable=False)
    cover_letter: Mapped[str] = mapped_column(Text, nullable=False, default="")
    cv_score: Mapped[float] = mapped_column(Float, nullable=False, default=0)
    exam_score: Mapped[float] = mapped_column(Float, nullable=False, default=0)
    interview_score: Mapped[float] = mapped_column(Float, nullable=False, default=0)
    final_score: Mapped[float] = mapped_column(Float, nullable=False, default=0)
    submitted_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow, index=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow, onupdate=datetime.utcnow)


class ParsedCV(Base):
    __tablename__ = "parsed_cvs"
    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    application_id: Mapped[int] = mapped_column(Integer, nullable=False, index=True, unique=True)
    extracted_text: Mapped[str] = mapped_column(Text, nullable=False, default="")
    extracted_skills: Mapped[str] = mapped_column(Text, nullable=False, default="")
    explanation: Mapped[str] = mapped_column(Text, nullable=False, default="")
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow, onupdate=datetime.utcnow)


class AuditLog(Base):
    __tablename__ = "audit_logs"
    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    user_id: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
    action: Mapped[str] = mapped_column(String(100), nullable=False, default="request")
    endpoint: Mapped[str] = mapped_column(String(255), nullable=False, default="")
    method: Mapped[str] = mapped_column(String(20), nullable=False, default="")
    ip_address: Mapped[str] = mapped_column(String(64), nullable=False, default="")
    details: Mapped[str] = mapped_column(Text, nullable=False, default="")
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=False), default=datetime.utcnow)


class SignupRequest(BaseModel):
    email: EmailStr
    password: str
    full_name: str = ""
    phone: str = ""
    nationality: str = ""
    date_of_birth: str = ""


class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class UpdateRoleRequest(BaseModel):
    role: str


class UpdateStatusRequest(BaseModel):
    status: str


class CreateJobRequest(BaseModel):
    title: str
    description: str
    required_skills: str
    qualifications: str = ""
    criteria_weights: str = ""
    deadline: str = ""


class UpdateJobRequest(BaseModel):
    title: str = ""
    description: str = ""
    required_skills: str = ""
    qualifications: str = ""
    criteria_weights: str = ""
    deadline: str = ""
    status: str = ""


class UserOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: int
    email: str
    full_name: str
    phone: str
    nationality: str
    date_of_birth: Optional[datetime]
    role: str
    status: str
    is_email_verified: bool
    is_phone_verified: bool
    failed_login_attempts: int
    locked_until: Optional[datetime]
    created_at: datetime
    updated_at: datetime


class AuthResponse(BaseModel):
    token: str
    user: UserOut


def parse_dob(raw: str) -> Optional[datetime]:
    if not raw:
        return None
    try:
        return datetime.strptime(raw, "%Y-%m-%d")
    except ValueError as exc:
        raise HTTPException(status_code=400, detail="date_of_birth must use YYYY-MM-DD format") from exc


def parse_deadline(raw: str) -> Optional[datetime]:
    if not raw:
        return None
    try:
        return datetime.strptime(raw, "%Y-%m-%d")
    except ValueError as exc:
        raise HTTPException(status_code=400, detail="deadline must use YYYY-MM-DD") from exc


def is_admin_role(role: str) -> bool:
    return role in {ROLE_ADMIN, ROLE_SUPER_ADMIN}


def is_strong_password(password: str) -> bool:
    return (
        bool(re.search(r"[A-Z]", password))
        and bool(re.search(r"[a-z]", password))
        and bool(re.search(r"[0-9]", password))
        and bool(re.search(r"[!@#$%^&*()\-_=+\[\]{}|;:,.<>?/`~]", password))
    )


def to_user_out(user: User) -> UserOut:
    return UserOut.model_validate(user)


def create_token(user: User) -> str:
    payload = {
        "user_id": user.id,
        "email": user.email,
        "role": user.role,
        "exp": datetime.now(timezone.utc) + timedelta(hours=JWT_TTL_HOURS),
    }
    return jwt.encode(payload, JWT_SECRET, algorithm=JWT_ALG)


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def get_current_user(request: Request, db: Session = Depends(get_db)) -> User:
    auth_header = request.headers.get("Authorization", "")
    parts = auth_header.split(" ")
    if len(parts) != 2 or parts[0] != "Bearer":
        raise HTTPException(status_code=401, detail="Invalid authorization header format")
    token = parts[1]
    try:
        payload = jwt.decode(token, JWT_SECRET, algorithms=[JWT_ALG])
        user_id = int(payload.get("user_id"))
    except (JWTError, TypeError, ValueError) as exc:
        raise HTTPException(status_code=401, detail="Invalid or expired token") from exc

    user = db.get(User, user_id)
    if not user or user.deleted_at is not None:
        raise HTTPException(status_code=401, detail="User not found")
    request.state.user = user
    return user


def require_recruiter_or_admin(user: User = Depends(get_current_user)) -> User:
    if user.role != ROLE_RECRUITER and not is_admin_role(user.role):
        raise HTTPException(status_code=403, detail="Recruiter or admin access required")
    return user


def require_admin(user: User = Depends(get_current_user)) -> User:
    if not is_admin_role(user.role):
        raise HTTPException(status_code=403, detail="Admin access required")
    return user


app = FastAPI(title="Imp101 FastAPI Backend")

origins = [os.getenv("CORS_ORIGIN", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

RATE_LIMIT_MAX = 100
_rate_limit_store: dict[str, list[float]] = {}


@app.middleware("http")
async def rate_limit_and_audit(request: Request, call_next):
    ip = request.client.host if request.client else "unknown"
    now = time.time()
    window_start = now - 60
    entries = [t for t in _rate_limit_store.get(ip, []) if t >= window_start]
    if len(entries) >= RATE_LIMIT_MAX:
        return JSONResponse(status_code=429, content={"error": "Too many requests"})
    entries.append(now)
    _rate_limit_store[ip] = entries

    response = await call_next(request)

    if request.method != "GET":
        db = SessionLocal()
        try:
            user_id = None
            user_obj = getattr(request.state, "user", None)
            if user_obj is not None:
                user_id = user_obj.id
            log = AuditLog(
                user_id=user_id,
                action="request",
                endpoint=request.url.path,
                method=request.method,
                ip_address=ip,
                details=json.dumps({"status": response.status_code}),
            )
            db.add(log)
            db.commit()
        except Exception:
            db.rollback()
        finally:
            db.close()
    return response


@app.on_event("startup")
def startup():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    try:
        admin = db.query(User).filter(User.email == "admin@admin.admin").first()
        pw_hash = pwd_context.hash("CqZP99nfbUI2M#3")
        if admin is None:
            admin = User(
                email="admin@admin.admin",
                password=pw_hash,
                full_name="System Super Admin",
                phone="",
                nationality="",
                role=ROLE_SUPER_ADMIN,
                status=STATUS_ACTIVE,
                is_email_verified=True,
                is_phone_verified=False,
            )
            db.add(admin)
        else:
            admin.role = ROLE_SUPER_ADMIN
            admin.status = STATUS_ACTIVE
            admin.password = pw_hash
        db.commit()
    finally:
        db.close()


@app.get("/healthz")
def health():
    return {"status": "ok"}


def register_public_routes(prefix: str):
    @app.post(f"{prefix}/signup", response_model=AuthResponse, status_code=201)
    def signup(req: SignupRequest, db: Session = Depends(get_db)):
        if not is_strong_password(req.password):
            raise HTTPException(status_code=400, detail="Password must include upper, lower, digit, and special character")

        existing = db.query(User).filter((User.email == req.email) | (User.phone == req.phone)).first()
        if existing:
            raise HTTPException(status_code=409, detail="User with this email or phone already exists")

        user = User(
            email=req.email,
            password=pwd_context.hash(req.password),
            full_name=req.full_name or "",
            phone=req.phone or "",
            nationality=req.nationality or "",
            date_of_birth=parse_dob(req.date_of_birth),
            role=ROLE_CANDIDATE,
            status=STATUS_ACTIVE,
            is_email_verified=False,
            is_phone_verified=False,
        )
        db.add(user)
        try:
            db.commit()
        except IntegrityError as exc:
            db.rollback()
            raise HTTPException(status_code=409, detail="User with this email or phone already exists") from exc
        db.refresh(user)
        return AuthResponse(token=create_token(user), user=to_user_out(user))

    @app.post(f"{prefix}/login", response_model=AuthResponse)
    def login(req: LoginRequest, db: Session = Depends(get_db)):
        user = db.query(User).filter(User.email == req.email).first()
        if not user:
            raise HTTPException(status_code=401, detail="Invalid email or password")

        if user.locked_until and user.locked_until > datetime.utcnow():
            raise HTTPException(status_code=401, detail="Account is temporarily locked due to failed login attempts")

        if user.status == STATUS_INACTIVE:
            raise HTTPException(status_code=403, detail="Account is inactive")

        if not pwd_context.verify(req.password, user.password):
            user.failed_login_attempts += 1
            if user.failed_login_attempts >= 5:
                user.locked_until = datetime.utcnow() + timedelta(minutes=15)
                user.status = STATUS_LOCKED
            db.commit()
            raise HTTPException(status_code=401, detail="Invalid email or password")

        user.failed_login_attempts = 0
        user.locked_until = None
        if user.status == STATUS_LOCKED:
            user.status = STATUS_ACTIVE
        db.commit()
        db.refresh(user)
        return AuthResponse(token=create_token(user), user=to_user_out(user))

    @app.get(f"{prefix}/jobs")
    def list_jobs(status: str = "", page: int = 1, limit: int = 20, db: Session = Depends(get_db)):
        if page < 1:
            page = 1
        if limit < 1 or limit > 100:
            limit = 20
        q = db.query(Job)
        if status:
            q = q.filter(Job.status == status)
        return q.order_by(Job.created_at.desc()).offset((page - 1) * limit).limit(limit).all()

    @app.get(f"{prefix}/jobs/{{job_id}}")
    def get_job(job_id: int, db: Session = Depends(get_db)):
        job = db.get(Job, job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        return job

    @app.get(f"{prefix}/users")
    def get_me(user: User = Depends(get_current_user)):
        return to_user_out(user)

    @app.get(f"{prefix}/users/{{user_id}}")
    def get_me_by_id(user_id: int, user: User = Depends(get_current_user)):
        if user.id != user_id:
            raise HTTPException(status_code=403, detail="Access denied")
        return to_user_out(user)

    @app.delete(f"{prefix}/users/me")
    def delete_my_data(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
        user.email = f"deleted_user_{user.id}@example.invalid"
        user.phone = f"deleted_{user.id}"
        user.full_name = "Deleted User"
        user.nationality = ""
        user.date_of_birth = None
        user.status = STATUS_INACTIVE
        user.is_email_verified = False
        user.is_phone_verified = False
        db.commit()
        return {"message": "Account anonymized"}

    @app.get(f"{prefix}/applications")
    def list_my_applications(user: User = Depends(get_current_user), db: Session = Depends(get_db)):
        return db.query(Application).filter(Application.candidate_id == user.id).order_by(Application.created_at.desc()).all()

    @app.get(f"{prefix}/applications/{{app_id}}")
    def get_my_application(app_id: int, user: User = Depends(get_current_user), db: Session = Depends(get_db)):
        app_item = db.query(Application).filter(Application.id == app_id, Application.candidate_id == user.id).first()
        if not app_item:
            raise HTTPException(status_code=404, detail="Application not found")
        return app_item

    @app.post(f"{prefix}/applications", status_code=202)
    def apply_to_job(
        job_id: int = Form(...),
        cover_letter: str = Form(""),
        cv: UploadFile = File(...),
        user: User = Depends(get_current_user),
        db: Session = Depends(get_db),
    ):
        if user.role != ROLE_CANDIDATE:
            raise HTTPException(status_code=403, detail="Candidate role required")
        job = db.get(Job, job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        if job.status != JOB_PUBLISHED:
            raise HTTPException(status_code=400, detail="Job is not accepting applications")
        if job.deadline and datetime.utcnow() > job.deadline:
            raise HTTPException(status_code=400, detail="Application deadline has passed")

        existing = db.query(Application).filter(Application.candidate_id == user.id, Application.job_id == job_id).first()
        if existing:
            raise HTTPException(status_code=409, detail="You already applied to this job")

        raw = cv.file.read()
        if len(raw) > 10 * 1024 * 1024:
            raise HTTPException(status_code=400, detail="CV file must be <= 10MB")

        ext = Path(cv.filename or "").suffix.lower()
        if ext not in {".pdf", ".docx", ".png", ".jpg", ".jpeg"}:
            raise HTTPException(status_code=400, detail="Invalid file type. Accepted: PDF, DOCX, PNG, JPG")

        upload_dir = Path("uploads") / "cv"
        upload_dir.mkdir(parents=True, exist_ok=True)
        save_name = f"{time.time_ns()}_{cv.filename}"
        save_path = upload_dir / save_name
        save_path.write_bytes(raw)

        app_item = Application(
            candidate_id=user.id,
            job_id=job_id,
            status=APP_SUBMITTED,
            cv_file_path=str(save_path),
            cover_letter=cover_letter,
            submitted_at=datetime.utcnow(),
        )
        db.add(app_item)
        db.commit()
        db.refresh(app_item)
        return app_item

    @app.post(f"{prefix}/jobs")
    def create_job(req: CreateJobRequest, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        job = Job(
            title=req.title,
            description=req.description,
            required_skills=req.required_skills,
            qualifications=req.qualifications,
            criteria_weights=req.criteria_weights,
            deadline=parse_deadline(req.deadline),
            status=JOB_DRAFT,
            created_by=user.id,
        )
        db.add(job)
        db.commit()
        db.refresh(job)
        return JSONResponse(status_code=201, content={
            "id": job.id,
            "title": job.title,
            "description": job.description,
            "required_skills": job.required_skills,
            "qualifications": job.qualifications,
            "criteria_weights": job.criteria_weights,
            "deadline": job.deadline.isoformat() if job.deadline else None,
            "status": job.status,
            "created_by": job.created_by,
            "created_at": job.created_at.isoformat(),
            "updated_at": job.updated_at.isoformat(),
        })

    @app.put(f"{prefix}/jobs/{{job_id}}")
    def update_job(job_id: int, req: UpdateJobRequest, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        job = db.get(Job, job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        if req.title:
            job.title = req.title
        if req.description:
            job.description = req.description
        if req.required_skills:
            job.required_skills = req.required_skills
        if req.qualifications:
            job.qualifications = req.qualifications
        if req.criteria_weights:
            job.criteria_weights = req.criteria_weights
        if req.status:
            job.status = req.status
        if req.deadline:
            job.deadline = parse_deadline(req.deadline)
        db.commit()
        db.refresh(job)
        return job

    @app.post(f"{prefix}/jobs/{{job_id}}/publish")
    def publish_job(job_id: int, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        job = db.get(Job, job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        job.status = JOB_PUBLISHED
        db.commit()
        return {"job_id": job_id, "status": JOB_PUBLISHED}

    @app.post(f"{prefix}/jobs/{{job_id}}/close")
    def close_job(job_id: int, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        job = db.get(Job, job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        job.status = JOB_CLOSED
        db.commit()
        return {"job_id": job_id, "status": JOB_CLOSED}

    @app.post(f"{prefix}/jobs/{{job_id}}/archive")
    def archive_job(job_id: int, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        job = db.get(Job, job_id)
        if not job:
            raise HTTPException(status_code=404, detail="Job not found")
        job.status = JOB_ARCHIVED
        db.commit()
        return {"job_id": job_id, "status": JOB_ARCHIVED}

    @app.get(f"{prefix}/job-rankings/{{job_id}}")
    def rank_candidates(job_id: int, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        apps = db.query(Application).filter(Application.job_id == job_id).order_by(Application.final_score.desc()).all()
        return apps

    @app.get(f"{prefix}/application-explainability/{{app_id}}")
    def app_explainability(app_id: int, user: User = Depends(require_recruiter_or_admin), db: Session = Depends(get_db)):
        parsed = db.query(ParsedCV).filter(ParsedCV.application_id == app_id).first()
        if not parsed:
            raise HTTPException(status_code=404, detail="Parsed analysis not found")
        return parsed

    @app.get(f"{prefix}/admin/users")
    def all_users(page: int = 1, limit: int = 20, user: User = Depends(require_admin), db: Session = Depends(get_db)):
        if page < 1:
            page = 1
        if limit < 1 or limit > 100:
            limit = 20
        records = db.query(User).order_by(User.created_at.desc()).offset((page - 1) * limit).limit(limit).all()
        return [to_user_out(u) for u in records]

    @app.patch(f"{prefix}/admin/users/{{user_id}}/role")
    def update_role(user_id: int, req: UpdateRoleRequest, user: User = Depends(require_admin), db: Session = Depends(get_db)):
        if req.role not in {ROLE_CANDIDATE, ROLE_RECRUITER, ROLE_ADMIN, ROLE_SUPER_ADMIN}:
            raise HTTPException(status_code=400, detail="Invalid role value")
        target = db.get(User, user_id)
        if not target:
            raise HTTPException(status_code=404, detail="User not found")
        target.role = req.role
        db.commit()
        return {"id": user_id, "role": req.role}

    @app.patch(f"{prefix}/admin/users/{{user_id}}/status")
    def update_status(user_id: int, req: UpdateStatusRequest, user: User = Depends(require_admin), db: Session = Depends(get_db)):
        if req.status not in {STATUS_ACTIVE, STATUS_INACTIVE, STATUS_LOCKED}:
            raise HTTPException(status_code=400, detail="Invalid status value")
        target = db.get(User, user_id)
        if not target:
            raise HTTPException(status_code=404, detail="User not found")
        target.status = req.status
        db.commit()
        return {"id": user_id, "status": req.status}


register_public_routes("")
register_public_routes("/api/v1")
