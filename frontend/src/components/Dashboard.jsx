import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import api from '../services/api'
import { removeToken, getUserRole } from '../services/auth'
import './Dashboard.css'

function Dashboard({ setIsAuthenticated }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [userRole, setUserRole] = useState(null)

  useEffect(() => {
    fetchUser()
    // Get role from token
    const role = getUserRole()
    setUserRole(role)
  }, [])

  const fetchUser = async () => {
    try {
      const response = await api.get('/users')
      setUser(response.data)
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
              <span className={`role-badge role-badge-${userRole || user?.role || 'regular'}`}>
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

        {userRole === 'super_admin' && (
          <div className="admin-section">
            <Link to="/admin" className="admin-panel-btn">
              Admin Panel
            </Link>
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
