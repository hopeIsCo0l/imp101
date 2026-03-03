export const setToken = (token) => {
  localStorage.setItem('token', token)
}

export const getToken = () => {
  return localStorage.getItem('token')
}

export const removeToken = () => {
  localStorage.removeItem('token')
}

// Decode JWT token and extract user information
export const decodeToken = (token) => {
  try {
    if (!token) return null
    
    // JWT tokens have 3 parts separated by dots: header.payload.signature
    const parts = token.split('.')
    if (parts.length !== 3) return null
    
    // Decode the payload (second part)
    const payload = parts[1]
    // Base64 decode and parse JSON
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    
    return decoded
  } catch (error) {
    console.error('Error decoding token:', error)
    return null
  }
}

// Get user role from token
export const getUserRole = () => {
  const token = getToken()
  if (!token) return null
  
  const decoded = decodeToken(token)
  return decoded?.role || null
}

// Get user ID from token
export const getUserId = () => {
  const token = getToken()
  if (!token) return null
  
  const decoded = decodeToken(token)
  return decoded?.user_id || null
}

export const getHomeRouteByRole = (role) => {
  if (!role) return '/login'
  return '/dashboard'
}
