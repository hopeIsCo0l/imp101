import { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import Login from './components/Login'
import Signup from './components/Signup'
import CandidatePortal from './components/CandidatePortal'
import RecruiterPortal from './components/RecruiterPortal'
import AdminPanel from './components/AdminPanel'
import { getHomeRouteByRole, getToken, getUserRole } from './services/auth'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [role, setRole] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = getToken()
    setIsAuthenticated(!!token)
    setRole(getUserRole())
    setLoading(false)
  }, [])

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  return (
    <Router>
      <div className="app">
        <Routes>
          <Route 
            path="/login" 
            element={
              isAuthenticated ? <Navigate to={getHomeRouteByRole(role)} /> : <Login setIsAuthenticated={setIsAuthenticated} />
            } 
          />
          <Route 
            path="/signup" 
            element={
              isAuthenticated ? <Navigate to={getHomeRouteByRole(role)} /> : <Signup setIsAuthenticated={setIsAuthenticated} />
            } 
          />
          <Route 
            path="/candidate" 
            element={
              isAuthenticated && role === 'candidate' ? <CandidatePortal setIsAuthenticated={setIsAuthenticated} /> : <Navigate to="/login" />
            } 
          />
          <Route 
            path="/recruiter" 
            element={
              isAuthenticated && role === 'recruiter' ? <RecruiterPortal setIsAuthenticated={setIsAuthenticated} /> : <Navigate to="/login" />
            } 
          />
          <Route 
            path="/admin" 
            element={
              isAuthenticated && (role === 'super_admin' || role === 'administrator') ? (
                <AdminPanel setIsAuthenticated={setIsAuthenticated} />
              ) : (
                <Navigate to="/login" />
              )
            } 
          />
          <Route path="/" element={<Navigate to={isAuthenticated ? getHomeRouteByRole(role) : "/login"} />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App
