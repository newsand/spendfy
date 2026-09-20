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

  getDelta(skuId, fonte, options = {}) {
    let params = `fonte=${fonte}`
    if (options.url) {
      params += `&url=${encodeURIComponent(options.url)}`
    } else if (options.monitor_id) {
      params += `&monitor_id=${options.monitor_id}`
    }
    return request(`/skus/${skuId}/observacoes/delta?${params}`)
  },

  getMonitor(skuId, options = {}) {
    let path = `/skus/${skuId}/monitor`
    const params = []
    if (options.url) {
      params.push(`url=${encodeURIComponent(options.url)}`)
    }
    if (options.monitor_id) {
      params.push(`monitor_id=${options.monitor_id}`)
    }
    if (params.length > 0) {
      path += `?${params.join('&')}`
    }
    return request(path).catch(e => {
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

  archiveMonitor(skuId, monitorId) {
    return request(`/skus/${skuId}/monitor/${monitorId}`, { method: 'DELETE' })
  },

  listMonitorHistory(skuId) {
    return request(`/skus/${skuId}/monitor/history`)
  },

  listTrackedSKUs() {
    return request('/dashboard/tracked-skus')
  }
}
