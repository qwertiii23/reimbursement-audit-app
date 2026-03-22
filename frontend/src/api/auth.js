import request from '@/utils/request'

export const login = (username, password) => {
  return request({
    url: '/auth/login',
    method: 'post',
    data: { username, password }
  })
}
