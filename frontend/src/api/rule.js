import request from '@/utils/request'

export const getRules = (params = {}) => {
  return request({
    url: '/rules',
    method: 'get',
    params
  })
}

export const createRule = (data) => {
  return request({
    url: '/rules',
    method: 'post',
    data
  })
}

export const updateRule = (id, data) => {
  return request({
    url: `/rules/${id}`,
    method: 'put',
    data
  })
}

export const deleteRule = (id) => {
  return request({
    url: `/rules/${id}`,
    method: 'delete'
  })
}

export const enableRule = (id) => {
  return request({
    url: `/rules/${id}/enable`,
    method: 'patch'
  })
}

export const disableRule = (id) => {
  return request({
    url: `/rules/${id}/disable`,
    method: 'patch'
  })
}

export const testRule = (id, data) => {
  return request({
    url: `/rules/${id}/test`,
    method: 'post',
    data
  })
}
