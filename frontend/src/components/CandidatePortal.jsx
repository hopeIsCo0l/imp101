import { useEffect, useState } from 'react'
import api from '../services/api'
import { removeToken } from '../services/auth'
import './Dashboard.css'

function CandidatePortal({ setIsAuthenticated }) {
  const [jobs, setJobs] = useState([])
  const [applications, setApplications] = useState([])
  const [selectedJobId, setSelectedJobId] = useState('')
  const [cvFile, setCvFile] = useState(null)
  const [coverLetter, setCoverLetter] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    loadJobs()
    loadApplications()
  }, [])

  const loadJobs = async () => {
    try {
      const { data } = await api.get('/jobs?status=published')
      setJobs(data)
    } catch (e) {
      setError(e.response?.data?.error || 'Failed to load jobs')
    }
  }

  const loadApplications = async () => {
    try {
      const { data } = await api.get('/applications')
      setApplications(data)
    } catch (e) {
      setError(e.response?.data?.error || 'Failed to load applications')
    }
  }

  const handleLogout = () => {
    removeToken()
    setIsAuthenticated(false)
    window.location.href = '/login'
  }

  const submitApplication = async (e) => {
    e.preventDefault()
    setError('')
    setMessage('')
    if (!selectedJobId || !cvFile) {
      setError('Please select a job and attach a CV file.')
      return
    }
    try {
      const formData = new FormData()
      formData.append('job_id', selectedJobId)
      formData.append('cover_letter', coverLetter)
      formData.append('cv', cvFile)
      await api.post('/applications', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setMessage('Application submitted and queued for parsing.')
      setCvFile(null)
      setCoverLetter('')
      loadApplications()
    } catch (e2) {
      setError(e2.response?.data?.error || 'Failed to submit application')
    }
  }

  return (
    <div className="dashboard-container">
      <div className="dashboard-card" style={{ maxWidth: 900 }}>
        <div className="dashboard-header">
          <h1>Candidate Portal</h1>
          <button onClick={handleLogout} className="logout-btn">Logout</button>
        </div>

        {message && <div className="info-success">{message}</div>}
        {error && <div className="error-message">{error}</div>}

        <div className="user-info">
          <h2>Apply to a Job</h2>
          <form onSubmit={submitApplication}>
            <div className="form-group">
              <label htmlFor="job">Job</label>
              <select id="job" value={selectedJobId} onChange={(e) => setSelectedJobId(e.target.value)}>
                <option value="">Select a published job</option>
                {jobs.map((job) => (
                  <option key={job.id} value={job.id}>{job.title}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label htmlFor="cv">CV (PDF, DOCX, JPG, PNG)</label>
              <input id="cv" type="file" onChange={(e) => setCvFile(e.target.files?.[0] || null)} />
            </div>
            <div className="form-group">
              <label htmlFor="coverLetter">Cover Letter</label>
              <textarea
                id="coverLetter"
                value={coverLetter}
                onChange={(e) => setCoverLetter(e.target.value)}
                placeholder="Optional cover letter"
              />
            </div>
            <button type="submit" className="admin-panel-btn">Submit Application</button>
          </form>
        </div>

        <div className="user-info">
          <h2>My Applications</h2>
          {applications.length === 0 ? (
            <p>No applications yet.</p>
          ) : (
            applications.map((app) => (
              <div className="info-item" key={app.id}>
                <span className="info-label">Job #{app.job_id}</span>
                <span className="info-value">
                  {app.status} | CV: {app.cv_score?.toFixed?.(2) ?? app.cv_score}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

export default CandidatePortal
