import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '../services/api'
import { getUserRole, removeToken } from '../services/auth'
import './AdminPanel.css'

function AdminPanel({ setIsAuthenticated }) {
  const [users, setUsers] = useState([])
  const [page, setPage] = useState(1)
  const [limit] = useState(20)
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState('all')
  const [loading, setLoading] = useState(true)
  const [actionLoadingId, setActionLoadingId] = useState(null)
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

    fetchUsers(page)
  }, [navigate, page])

  const fetchUsers = async (nextPage = 1) => {
    try {
      const response = await api.get(`/admin/users?page=${nextPage}&limit=${limit}`)
      setUsers(response.data)
      setDebugInfo(null) // Clear debug info on success
      setError('')
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

  const handleLogout = () => {
    removeToken()
    if (setIsAuthenticated) setIsAuthenticated(false)
    window.location.href = '/login'
  }

  const updateUserRole = async (userId, role) => {
    try {
      setActionLoadingId(userId)
      await api.patch(`/admin/users/${userId}/role`, { role })
      setUsers((prev) => prev.map((user) => (user.id === userId ? { ...user, role } : user)))
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update role')
    } finally {
      setActionLoadingId(null)
    }
  }

  const updateUserStatus = async (userId, status) => {
    try {
      setActionLoadingId(userId)
      await api.patch(`/admin/users/${userId}/status`, { status })
      setUsers((prev) => prev.map((user) => (user.id === userId ? { ...user, status } : user)))
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update status')
    } finally {
      setActionLoadingId(null)
    }
  }

  const visibleUsers = users.filter((user) => {
    const email = String(user.email || '').toLowerCase()
    const role = String(user.role || '')
    const status = String(user.status || '')
    const matchSearch = email.includes(search.toLowerCase())
    const matchRole = roleFilter === 'all' ? true : role === roleFilter
    const matchStatus = statusFilter === 'all' ? true : status === statusFilter
    return matchSearch && matchRole && matchStatus
  })

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
              onClick={handleLogout} 
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
          <div style={{ display: 'flex', gap: '8px' }}>
            <button onClick={() => navigate('/dashboard')} className="back-btn">
              Dashboard
            </button>
            <button onClick={handleLogout} className="back-btn" style={{ backgroundColor: '#dc3545' }}>
              Logout
            </button>
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}

        <div className="admin-toolbar">
          <div className="toolbar-item">
            <label htmlFor="searchEmail">Search email</label>
            <input
              id="searchEmail"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="user@example.com"
            />
          </div>
          <div className="toolbar-item">
            <label htmlFor="roleFilter">Role</label>
            <select id="roleFilter" value={roleFilter} onChange={(e) => setRoleFilter(e.target.value)}>
              <option value="all">All</option>
              <option value="candidate">candidate</option>
              <option value="recruiter">recruiter</option>
              <option value="administrator">administrator</option>
              <option value="super_admin">super_admin</option>
            </select>
          </div>
          <div className="toolbar-item">
            <label htmlFor="statusFilter">Status</label>
            <select id="statusFilter" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              <option value="all">All</option>
              <option value="active">active</option>
              <option value="inactive">inactive</option>
              <option value="locked">locked</option>
            </select>
          </div>
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
                  <th>Status</th>
                  <th>Created At</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {visibleUsers.length === 0 ? (
                  <tr>
                    <td colSpan="6" className="no-data">
                      No users found
                    </td>
                  </tr>
                ) : (
                  visibleUsers.map((user) => (
                    <tr key={user.id}>
                      <td>{user.id}</td>
                      <td>{user.email}</td>
                      <td>
                        {user.role || 'candidate'}
                      </td>
                      <td>
                        {user.status || 'active'}
                      </td>
                      <td>
                        {user.created_at
                          ? new Date(user.created_at).toLocaleString()
                          : 'N/A'}
                      </td>
                      <td>
                        <div className="action-wrap">
                          <select
                            value={user.role || 'candidate'}
                            onChange={(e) => updateUserRole(user.id, e.target.value)}
                            disabled={actionLoadingId === user.id}
                          >
                            <option value="candidate">candidate</option>
                            <option value="recruiter">recruiter</option>
                            <option value="administrator">administrator</option>
                            <option value="super_admin">super_admin</option>
                          </select>
                          <select
                            value={user.status || 'active'}
                            onChange={(e) => updateUserStatus(user.id, e.target.value)}
                            disabled={actionLoadingId === user.id}
                          >
                            <option value="active">active</option>
                            <option value="inactive">inactive</option>
                            <option value="locked">locked</option>
                          </select>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
          <div className="pagination-row">
            <button className="back-btn" disabled={page === 1} onClick={() => setPage((prev) => Math.max(1, prev - 1))}>
              Previous
            </button>
            <span>Page {page}</span>
            <button className="back-btn" disabled={users.length < limit} onClick={() => setPage((prev) => prev + 1)}>
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default AdminPanel
