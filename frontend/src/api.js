const API_BASE = '/api/v1'

async function request(path, options = {}) {
  const url = `${API_BASE}${path}`
  const res = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }

  if (res.status === 204) return null
  return res.json()
}

export const api = {
  listSKUs() {
    return request('/skus')
  },

  getSKU(id) {
    return request(`/skus/${id}`)
  },

  createSKU(data) {
    return request('/skus', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  updateSKU(id, data) {
    return request(`/skus/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  deleteSKU(id) {
    return request(`/skus/${id}`, { method: 'DELETE' })
  },

  listObservacoes(skuId, fonte = null) {
    const params = fonte ? `?fonte=${fonte}` : ''
    return request(`/skus/${skuId}/observacoes${params}`)
  },

  createObservacao(data) {
    return request('/observacoes', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  getDelta(skuId, fonte) {
    return request(`/skus/${skuId}/observacoes/delta?fonte=${fonte}`)
  },

  getMonitor(skuId) {
    return request(`/skus/${skuId}/monitor`).catch(e => {
      if (e.message === 'no active monitor') return null
      throw e
    })
  },

  setMonitor(skuId, data) {
    return request(`/skus/${skuId}/monitor`, {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  updateMonitorLimiar(skuId, data) {
    return request(`/skus/${skuId}/monitor/limiar`, {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  listMonitorHistory(skuId) {
    return request(`/skus/${skuId}/monitor/history`)
  },

  listTrackedSKUs() {
    return request('/dashboard/tracked-skus')
  }
}
