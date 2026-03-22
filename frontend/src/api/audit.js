import request from '@/utils/request'

export const startAudit = (data) => {
  return request({
    url: '/audit',
    method: 'post',
    data
  })
}

export const getAuditStatus = (auditId) => {
  return request({
    url: `/audit/${auditId}/status`,
    method: 'get'
  })
}

export const getAuditResult = (auditId) => {
  return request({
    url: `/audit/${auditId}/result`,
    method: 'get'
  })
}

export const retryAudit = (auditId) => {
  return request({
    url: `/audit/${auditId}/retry`,
    method: 'post'
  })
}

export const manualAudit = (auditId, action, reason) => {
  return request({
    url: `/audit/${auditId}/manual-audit`,
    method: 'post',
    data: { action, reason }
  })
}

export const getFlowLogsByReimbursementId = (reimbursementId) => {
  return request({
    url: `/audit/flow-logs/reimbursement/${reimbursementId}`,
    method: 'get'
  })
}

export const getFlowLogsByAuditId = (auditId) => {
  return request({
    url: `/audit/flow-logs/audit/${auditId}`,
    method: 'get'
  })
}

export const getFlowLogs = (auditId) => {
  return request({
    url: '/audit/flow-logs',
    method: 'get',
    params: { audit_id: auditId }
  })
}

export const withdrawAudit = (reimbursementId) => {
  return request({
    url: `/audit/${reimbursementId}/withdraw`,
    method: 'post'
  })
}
