import service from '~/utils/request'

export const register = (data) => {
  return service({
    url: '/api/v1/user/register',
    method: 'post',
    data,
  })
}

export const login = (data) => {
  return service({
    url: '/api/v1/user/login',
    method: 'post',
    data,
  })
}

export const getUser = (params) => {
  return service({
    url: '/api/v1/user',
    method: 'get',
    params,
  })
}

export const updateUserPassword = (data) => {
  return service({
    url: '/api/v1/user/password',
    method: 'put',
    data,
  })
}

export const updateUser = (data) => {
  return service({
    url: '/api/v1/user/password',
    method: 'put',
    data,
  })
}

export const deleteUser = (params) => {
  return service({
    url: '/api/v1/user',
    method: 'delete',
    params,
  })
}

export const listUser = (params) => {
  return service({
    url: '/api/v1/user/list',
    method: 'get',
    params,
  })
}

export const getUserCaptcha = (params) => {
  return service({
    url: '/api/v1/user/captcha',
    method: 'get',
    params,
  })
}

