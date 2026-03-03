import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import api from '../services/api'
import { removeToken, getUserRole } from '../services/auth'
import './Dashboard.css'

function Dashboard({ setIsAuthenticated }) {
  const [user, setUser] = useState(null)
  const [jobs, setJobs] = useState([])
  const [applications, setApplications] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [userRole, setUserRole] = useState(null)

  useEffect(() => {
    const role = getUserRole()
    setUserRole(role)
    loadDashboard(role)
  }, [])

  const loadDashboard = async (role) => {
    try {
      const userPromise = api.get('/users')
      const jobsPromise = role === 'candidate' ? api.get('/jobs?status=published') : api.get('/jobs')
      const applicationsPromise = role === 'candidate' ? api.get('/applications') : Promise.resolve({ data: [] })
      const [userResponse, jobsResponse, applicationsResponse] = await Promise.all([userPromise, jobsPromise, applicationsPromise])
      setUser(userResponse.data)
      setJobs(Array.isArray(jobsResponse.data) ? jobsResponse.data : [])
      setApplications(Array.isArray(applicationsResponse.data) ? applicationsResponse.data : [])
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to fetch user data')
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    removeToken()
    setIsAuthenticated(false)
    window.location.href = '/login'
  }

  const statusCount = (status) => jobs.filter((job) => job.status === status).length
  const roleHomeRoute = userRole === 'candidate' ? '/candidate' : userRole === 'recruiter' ? '/recruiter' : '/admin'

  if (loading) {
    return (
      <div className="dashboard-container">
        <div className="dashboard-card">
          <div className="loading">Loading...</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="dashboard-container">
        <div className="dashboard-card">
          <div className="error-message">{error}</div>
          <button onClick={handleLogout} className="logout-btn">
            Go to Login
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="dashboard-container">
      <div className="dashboard-card">
        <div className="dashboard-header">
          <h1>Dashboard</h1>
          <button onClick={handleLogout} className="logout-btn">
            Logout
          </button>
        </div>

        <div className="stats-grid">
          <div className="stat-card">
            <p className="stat-label">Active Role</p>
            <p className="stat-value">{userRole || '-'}</p>
          </div>
          <div className="stat-card">
            <p className="stat-label">Jobs Visible</p>
            <p className="stat-value">{jobs.length}</p>
          </div>
          {userRole === 'candidate' && (
            <div className="stat-card">
              <p className="stat-label">My Applications</p>
              <p className="stat-value">{applications.length}</p>
            </div>
          )}
          {(userRole === 'recruiter' || userRole === 'administrator' || userRole === 'super_admin') && (
            <>
              <div className="stat-card">
                <p className="stat-label">Published Jobs</p>
                <p className="stat-value">{statusCount('published')}</p>
              </div>
              <div className="stat-card">
                <p className="stat-label">Draft Jobs</p>
                <p className="stat-value">{statusCount('draft')}</p>
              </div>
            </>
          )}
        </div>

        <div className="user-info">
          <h2>User Information</h2>
          <div className="info-item">
            <span className="info-label">ID:</span>
            <span className="info-value">{user?.id}</span>
          </div>
          <div className="info-item">
            <span className="info-label">Email:</span>
            <span className="info-value">{user?.email}</span>
          </div>
          <div className="info-item">
            <span className="info-label">Role:</span>
            <span className="info-value">
              <span className={`role-badge role-badge-${userRole || user?.role || 'candidate'}`}>
                {userRole || user?.role || 'regular'}
              </span>
            </span>
          </div>
          <div className="info-item">
            <span className="info-label">Created At:</span>
            <span className="info-value">
              {user?.created_at ? new Date(user.created_at).toLocaleString() : 'N/A'}
            </span>
          </div>
          <div className="info-item">
            <span className="info-label">Updated At:</span>
            <span className="info-value">
              {user?.updated_at ? new Date(user.updated_at).toLocaleString() : 'N/A'}
            </span>
          </div>
        </div>

        <div className="admin-section">
          <h2>Quick Actions</h2>
          <div className="actions-row">
            <Link to={roleHomeRoute} className="admin-panel-btn">
              Open {userRole === 'candidate' ? 'Candidate Portal' : userRole === 'recruiter' ? 'Recruiter Portal' : 'Admin Panel'}
            </Link>
            {userRole === 'candidate' && (
              <Link to="/candidate" className="secondary-btn">
                Apply to Jobs
              </Link>
            )}
            {(userRole === 'administrator' || userRole === 'super_admin') && (
              <Link to="/admin" className="secondary-btn">
                Manage Users
              </Link>
            )}
          </div>
        </div>

        {userRole === 'candidate' && (
          <div className="user-info">
            <h2>Application Activity</h2>
            {applications.length === 0 ? (
              <p>No applications submitted yet.</p>
            ) : (
              applications.slice(0, 5).map((app) => (
                <div className="info-item" key={app.id}>
                  <span className="info-label">Application #{app.id}</span>
                  <span className="info-value">{app.status} | Final Score: {Number(app.final_score || 0).toFixed(2)}</span>
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
