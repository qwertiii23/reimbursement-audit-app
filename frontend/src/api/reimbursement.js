import request from '@/utils/request'

export const getReimbursementsByUser = (userId, page = 1, pageSize = 10, params = {}) => {
  return request({
    url: '/reimbursements/user',
    method: 'get',
    params: { user_id: userId, page, page_size: pageSize, ...params }
  })
}

export const getAllReimbursements = (page = 1, pageSize = 10, params = {}) => {
  return request({
    url: '/reimbursements/all',
    method: 'get',
    params: { page, page_size: pageSize, ...params }
  })
}

export const getReimbursementById = (id) => {
  return request({
    url: `/reimbursement/${id}`,
    method: 'get'
  })
}

export const updateReimbursement = (id, data) => {
  return request({
    url: `/reimbursement/${id}`,
    method: 'put',
    data
  })
}

export const uploadReimbursement = (formData) => {
  return request({
    url: '/reimbursement/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const uploadInvoice = (reimbursementId, category, file) => {
  const formData = new FormData()
  formData.append('reimbursement_id', reimbursementId)
  formData.append('category', category)
  formData.append('invoice', file)
  return request({
    url: '/invoices/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const batchUploadInvoices = (reimbursementId, category, files) => {
  const formData = new FormData()
  formData.append('reimbursement_id', reimbursementId)
  formData.append('category', category)
  files.forEach((file, index) => {
    formData.append(`files[${index}]`, file)
  })
  return request({
    url: '/invoices/batch-upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const triggerOCR = (invoiceId) => {
  return request({
    url: '/invoices/ocr',
    method: 'post',
    data: { invoice_id: invoiceId }
  })
}

export const updateInvoiceImage = (invoiceId, file) => {
  const formData = new FormData()
  formData.append('invoice_id', invoiceId)
  formData.append('file', file)
  return request({
    url: '/invoices/update-image',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const getAuditReport = (reimbursementId) => {
  return request({
    url: `/reimbursement/${reimbursementId}/audit-report`,
    method: 'get'
  })
}
