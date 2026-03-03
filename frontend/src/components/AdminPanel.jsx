import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../services/api'
import { getUserRole, removeToken } from '../services/auth'
import './AdminPanel.css'

function AdminPanel({ setIsAuthenticated }) {
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [debugInfo, setDebugInfo] = useState(null)
  const navigate = useNavigate()

  useEffect(() => {
    // Check if user is admin
    const role = getUserRole()
    if (!role) {
      setError('Unable to determine user role. Please log out and log back in to refresh your session.')
      setLoading(false)
      return
    }
    if (role !== 'super_admin' && role !== 'administrator') {
      setError('Access denied. You do not have permission to access the admin panel.')
      setLoading(false)
      return
    }

    fetchUsers()
  }, [navigate])

  const fetchUsers = async () => {
    try {
      const response = await api.get('/admin/users')
      setUsers(response.data)
      setDebugInfo(null) // Clear debug info on success
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.message || 'Failed to fetch users'
      const status = err.response?.status
      const requestUrl = err.config?.url || err.request?.responseURL || 'Unknown URL'
      const baseURL = err.config?.baseURL || 'Unknown base URL'
      const fullUrl = baseURL === '/api' ? `${baseURL}${requestUrl}` : `${baseURL}${requestUrl.startsWith('/') ? '' : '/'}${requestUrl}`
      
      // Prepare debug info for 404 errors
      const debug = {
        fullUrl,
        baseURL,
        requestPath: requestUrl,
        status,
        message: errorMessage,
        method: err.config?.method?.toUpperCase() || 'GET',
        responseData: err.response?.data,
      }
      
      if (status === 401) {
        setError('Authentication failed. Please log out and log back in.')
        setDebugInfo(null)
      } else if (status === 403) {
        setError('Access denied. You do not have permission to view this page.')
        setDebugInfo(null)
      } else if (status === 404) {
        setError('404 Not Found: The requested endpoint was not found.')
        setDebugInfo(debug)
      } else if (status >= 500) {
        setError('Server error. Please try again later.')
        setDebugInfo(debug)
      } else {
        setError(`Error: ${errorMessage}`)
        setDebugInfo(debug)
      }
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="admin-container">
        <div className="admin-card">
          <div className="loading">Loading...</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="admin-container">
        <div className="admin-card">
          <div className="error-message">
            <strong>Error:</strong> {error}
          </div>
          {debugInfo && (
            <div style={{ 
              marginTop: '1rem', 
              padding: '1rem', 
              backgroundColor: '#f8f9fa', 
              border: '1px solid #dee2e6', 
              borderRadius: '4px',
              fontSize: '0.9rem'
            }}>
              <strong>Debug Information:</strong>
              <pre style={{ 
                marginTop: '0.5rem', 
                padding: '0.5rem', 
                backgroundColor: '#fff', 
                border: '1px solid #dee2e6',
                borderRadius: '4px',
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word'
              }}>
                <strong>Full URL:</strong> {debugInfo.fullUrl}{'\n'}
                <strong>Base URL:</strong> {debugInfo.baseURL}{'\n'}
                <strong>Request Path:</strong> {debugInfo.requestPath}{'\n'}
                <strong>Method:</strong> {debugInfo.method}{'\n'}
                <strong>Status:</strong> {debugInfo.status}{'\n'}
                <strong>Message:</strong> {debugInfo.message}{'\n'}
                {debugInfo.responseData && (
                  <>
                    <strong>Response Data:</strong>{'\n'}
                    {JSON.stringify(debugInfo.responseData, null, 2)}
                  </>
                )}
              </pre>
              <div style={{ marginTop: '0.5rem', fontSize: '0.85rem', color: '#6c757d' }}>
                <strong>Please check:</strong>
                <ul style={{ marginTop: '0.25rem', marginBottom: 0 }}>
                  <li>Is the backend server running on port 8080?</li>
                  <li>Is the Vite proxy configured correctly in vite.config.js?</li>
                  <li>Does the endpoint '{debugInfo.requestPath}' exist on the backend?</li>
                  <li>Check browser console and network tab for more details</li>
                </ul>
              </div>
            </div>
          )}
          <div style={{ marginTop: '1rem', display: 'flex', gap: '1rem' }}>
            <button onClick={() => navigate('/dashboard')} className="back-btn">
              Back to Dashboard
            </button>
            <button 
              onClick={() => {
                // Clear token and redirect to login
                removeToken()
                if (setIsAuthenticated) setIsAuthenticated(false)
                window.location.href = '/login'
              }} 
              className="back-btn"
              style={{ backgroundColor: '#dc3545' }}
            >
              Log Out
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="admin-container">
      <div className="admin-card">
        <div className="admin-header">
          <h1>Admin Panel</h1>
          <button onClick={() => navigate('/dashboard')} className="back-btn">
            Back to Dashboard
          </button>
        </div>

        <div className="admin-content">
          <h2>All Users</h2>
          <div className="users-table-container">
            <table className="users-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Created At</th>
                </tr>
              </thead>
              <tbody>
                {users.length === 0 ? (
                  <tr>
                    <td colSpan="4" className="no-data">
                      No users found
                    </td>
                  </tr>
                ) : (
                  users.map((user) => (
                    <tr key={user.id}>
                      <td>{user.id}</td>
                      <td>{user.email}</td>
                      <td>
                        <span className={`role-badge role-badge-${user.role || 'regular'}`}>
                          {user.role || 'regular'}
                        </span>
                      </td>
                      <td>
                        {user.created_at
                          ? new Date(user.created_at).toLocaleString()
                          : 'N/A'}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}

export default AdminPanel
