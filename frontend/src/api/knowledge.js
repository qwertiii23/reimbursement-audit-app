import request from '@/utils/request'

export const getKnowledgeFiles = (params = {}) => {
  return request({
    url: '/knowledge/files',
    method: 'get',
    params
  })
}

export const uploadKnowledgeFile = (data) => {
  return request({
    url: '/knowledge/upload',
    method: 'post',
    data
  })
}

export const getKnowledgeFile = (id) => {
  return request({
    url: `/knowledge/files/${id}`,
    method: 'get'
  })
}

export const updateKnowledgeFile = (id, data) => {
  return request({
    url: `/knowledge/files/${id}`,
    method: 'put',
    data
  })
}

export const deleteKnowledgeFile = (id) => {
  return request({
    url: `/knowledge/files/${id}`,
    method: 'delete'
  })
}

export const downloadKnowledgeFile = (id) => {
  return request({
    url: `/knowledge/download/${id}`,
    method: 'get',
    responseType: 'blob'
  })
}

export const viewKnowledgeFile = (id) => {
  return request({
    url: `/knowledge/view/${id}`,
    method: 'get'
  })
}
