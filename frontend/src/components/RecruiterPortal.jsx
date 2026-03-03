import { useEffect, useState } from 'react'
import api from '../services/api'
import { removeToken } from '../services/auth'
import './Dashboard.css'

const emptyJob = {
  title: '',
  description: '',
  required_skills: '',
  qualifications: '',
  criteria_weights: '{"cv":40,"exam":35,"interview":25}',
  deadline: '',
}

function RecruiterPortal({ setIsAuthenticated }) {
  const [jobs, setJobs] = useState([])
  const [jobForm, setJobForm] = useState(emptyJob)
  const [selectedJobId, setSelectedJobId] = useState('')
  const [ranking, setRanking] = useState([])
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  useEffect(() => {
    loadJobs()
  }, [])

  const loadJobs = async () => {
    try {
      const { data } = await api.get('/jobs')
      setJobs(data)
    } catch (e) {
      setError(e.response?.data?.error || 'Failed to load jobs')
    }
  }

  const handleLogout = () => {
    removeToken()
    setIsAuthenticated(false)
    window.location.href = '/login'
  }

  const createJob = async (e) => {
    e.preventDefault()
    setError('')
    setMessage('')
    try {
      await api.post('/jobs', jobForm)
      setMessage('Job created as draft.')
      setJobForm(emptyJob)
      loadJobs()
    } catch (e2) {
      setError(e2.response?.data?.error || 'Failed to create job')
    }
  }

  const publishJob = async (id) => {
    try {
      await api.post(`/jobs/${id}/publish`)
      loadJobs()
    } catch (e) {
      setError(e.response?.data?.error || 'Failed to publish job')
    }
  }

  const loadRanking = async () => {
    if (!selectedJobId) return
    try {
      const { data } = await api.get(`/job-rankings/${selectedJobId}`)
      setRanking(data)
    } catch (e) {
      setError(e.response?.data?.error || 'Failed to fetch ranking')
    }
  }

  return (
    <div className="dashboard-container">
      <div className="dashboard-card" style={{ maxWidth: 1000 }}>
        <div className="dashboard-header">
          <h1>Recruiter Portal</h1>
          <button onClick={handleLogout} className="logout-btn">Logout</button>
        </div>

        {message && <div className="info-success">{message}</div>}
        {error && <div className="error-message">{error}</div>}

        <div className="user-info">
          <h2>Create Job</h2>
          <form onSubmit={createJob}>
            <div className="form-group">
              <label htmlFor="title">Title</label>
              <input id="title" value={jobForm.title} onChange={(e) => setJobForm({ ...jobForm, title: e.target.value })} />
            </div>
            <div className="form-group">
              <label htmlFor="description">Description</label>
              <textarea id="description" value={jobForm.description} onChange={(e) => setJobForm({ ...jobForm, description: e.target.value })} />
            </div>
            <div className="form-group">
              <label htmlFor="skills">Required Skills (comma-separated)</label>
              <input id="skills" value={jobForm.required_skills} onChange={(e) => setJobForm({ ...jobForm, required_skills: e.target.value })} />
            </div>
            <div className="form-group">
              <label htmlFor="quals">Qualifications</label>
              <input id="quals" value={jobForm.qualifications} onChange={(e) => setJobForm({ ...jobForm, qualifications: e.target.value })} />
            </div>
            <div className="form-group">
              <label htmlFor="weights">Criteria Weights (json text)</label>
              <input id="weights" value={jobForm.criteria_weights} onChange={(e) => setJobForm({ ...jobForm, criteria_weights: e.target.value })} />
            </div>
            <div className="form-group">
              <label htmlFor="deadline">Deadline (YYYY-MM-DD)</label>
              <input id="deadline" value={jobForm.deadline} onChange={(e) => setJobForm({ ...jobForm, deadline: e.target.value })} />
            </div>
            <button type="submit" className="admin-panel-btn">Create Draft</button>
          </form>
        </div>

        <div className="user-info">
          <h2>Jobs</h2>
          {jobs.map((job) => (
            <div key={job.id} className="info-item">
              <span className="info-label">
                #{job.id} {job.title}
              </span>
              <span className="info-value">
                {job.status}{' '}
                {job.status !== 'published' && (
                  <button className="admin-panel-btn" style={{ marginLeft: 8 }} onClick={() => publishJob(job.id)}>
                    Publish
                  </button>
                )}
              </span>
            </div>
          ))}
        </div>

        <div className="user-info">
          <h2>Candidate Ranking</h2>
          <div className="form-group">
            <label htmlFor="jobRank">Select Job</label>
            <select id="jobRank" value={selectedJobId} onChange={(e) => setSelectedJobId(e.target.value)}>
              <option value="">Select job</option>
              {jobs.map((job) => (
                <option key={job.id} value={job.id}>
                  {job.title}
                </option>
              ))}
            </select>
          </div>
          <button className="admin-panel-btn" onClick={loadRanking}>Load Ranking</button>
          {ranking.map((item, idx) => (
            <div className="info-item" key={item.id}>
              <span className="info-label">#{idx + 1} App {item.id}</span>
              <span className="info-value">Final Score: {item.final_score}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default RecruiterPortal
