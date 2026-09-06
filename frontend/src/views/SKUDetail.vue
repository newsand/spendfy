<template>
  <div class="sku-detail" v-if="sku">
    <div class="header">
      <div>
        <h1>{{ sku.nome }}</h1>
        <p class="meta" v-if="sku.marca">{{ sku.marca }} {{ sku.tamanho_ou_variante }}</p>
        <span class="badge">{{ sku.unidade_padrao }}</span>
      </div>
      <div class="actions">
        <router-link :to="`/skus/${sku.id}/edit`" class="btn">Editar</router-link>
        <button @click="deleteSKU" class="btn btn-danger">Excluir</button>
      </div>
    </div>

    <div class="tabs">
      <button :class="{ active: activeTab === 'compra' }" @click="activeTab = 'compra'">
        Compras
      </button>
      <button :class="{ active: activeTab === 'rastreio' }" @click="activeTab = 'rastreio'">
        Rastreio
      </button>
      <button :class="{ active: activeTab === 'monitor' }" @click="activeTab = 'monitor'">
        Monitor
      </button>
    </div>

    <div v-if="activeTab === 'compra'" class="tab-content">
      <div class="delta-card" v-if="deltaCompra">
        <h3>Última observação</h3>
        <p class="price">R$ {{ formatPrice(deltaCompra.current_preco) }}</p>
        <p v-if="deltaCompra.has_previous" class="delta" :class="deltaClass(deltaCompra.delta_absoluto)">
          {{ formatDelta(deltaCompra) }}
        </p>
      </div>

      <h3>Registrar Compra</h3>
      <form @submit.prevent="submitCompra" class="form">
        <div class="form-row">
          <label>Preço unitário (R$)</label>
          <input v-model.number="compraForm.preco" type="number" step="0.01" required />
        </div>
        <div class="form-row">
          <label>Quantidade (opcional)</label>
          <input v-model.number="compraForm.quantidade" type="number" step="0.01" />
        </div>
        <div class="form-row">
          <label>Loja</label>
          <input v-model="compraForm.loja" type="text" required />
        </div>
        <div class="form-row">
          <label>Notas (opcional)</label>
          <textarea v-model="compraForm.notas"></textarea>
        </div>
        <button type="submit" class="btn">Salvar</button>
      </form>

      <h3>Histórico de Compras</h3>
      <table v-if="observacoesCompra.length" class="table">
        <thead>
          <tr>
            <th>Data</th>
            <th>Preço</th>
            <th>Qtd</th>
            <th>Loja</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="obs in observacoesCompra" :key="obs.id">
            <td>{{ formatDate(obs.data) }}</td>
            <td>R$ {{ formatPrice(obs.preco) }}</td>
            <td>{{ obs.quantidade || '-' }}</td>
            <td>{{ obs.loja }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">Nenhuma compra registrada.</p>
    </div>

    <div v-if="activeTab === 'rastreio'" class="tab-content">
      <div class="delta-card" v-if="deltaRastreio">
        <h3>Última coleta (URL ativa)</h3>
        <p class="price">R$ {{ formatPrice(deltaRastreio.current_preco) }}</p>
        <p v-if="deltaRastreio.has_previous" class="delta" :class="deltaClass(deltaRastreio.delta_absoluto)">
          {{ formatDelta(deltaRastreio) }}
        </p>
      </div>

      <h3>Histórico de Rastreio</h3>
      <table v-if="observacoesRastreio.length" class="table">
        <thead>
          <tr>
            <th>Data</th>
            <th>Preço</th>
            <th>Loja</th>
            <th>URL</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="obs in observacoesRastreio" :key="obs.id" :class="{ archived: isArchivedUrl(obs.url) }">
            <td>{{ formatDate(obs.data) }}</td>
            <td>R$ {{ formatPrice(obs.preco) }}</td>
            <td>{{ obs.loja }}</td>
            <td class="url">{{ obs.url }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">Nenhum rastreio realizado.</p>
    </div>

    <div v-if="activeTab === 'monitor'" class="tab-content">
      <div v-if="monitor" class="monitor-info">
        <h3>Monitor Ativo</h3>
        <p><strong>URL:</strong> <a :href="monitor.url" target="_blank">{{ monitor.url }}</a></p>
        <p v-if="monitor.limiar_modo">
          <strong>Limiar:</strong> 
          {{ monitor.limiar_modo === 'absoluto' ? `R$ ${formatPrice(monitor.limiar_valor)}` : `${monitor.limiar_valor}%` }}
          ({{ monitor.limiar_modo }})
        </p>
        <p v-else><em>Sem limiar de alerta configurado</em></p>
      </div>

      <h3>{{ monitor ? 'Trocar URL (arquiva atual)' : 'Configurar Monitor' }}</h3>
      <form @submit.prevent="submitMonitor" class="form">
        <div class="form-row">
          <label>URL para monitorar</label>
          <input v-model="monitorForm.url" type="url" required placeholder="https://..." />
        </div>
        <div class="form-row">
          <label>CSS Selector (opcional)</label>
          <input v-model="monitorForm.css_selector" type="text" placeholder=".price, #product-price" />
        </div>
        <div class="form-row">
          <label>Regex Pattern (opcional)</label>
          <input v-model="monitorForm.regex_pattern" type="text" placeholder="R\$\s*([\d.,]+)" />
        </div>
        <div class="form-row">
          <label>Limiar de Alerta</label>
          <select v-model="monitorForm.limiar_modo">
            <option :value="null">Sem alerta</option>
            <option value="absoluto">Absoluto (R$)</option>
            <option value="percentual">Percentual (%)</option>
          </select>
        </div>
        <div class="form-row" v-if="monitorForm.limiar_modo">
          <label>Valor do Limiar</label>
          <input v-model.number="monitorForm.limiar_valor" type="number" step="0.01" required />
        </div>
        <button type="submit" class="btn">{{ monitor ? 'Trocar URL' : 'Criar Monitor' }}</button>
      </form>

      <div v-if="monitor">
        <h3>Atualizar Limiar</h3>
        <form @submit.prevent="updateLimiar" class="form">
          <div class="form-row">
            <label>Modo</label>
            <select v-model="limiarForm.limiar_modo">
              <option :value="null">Sem alerta</option>
              <option value="absoluto">Absoluto (R$)</option>
              <option value="percentual">Percentual (%)</option>
            </select>
          </div>
          <div class="form-row" v-if="limiarForm.limiar_modo">
            <label>Valor</label>
            <input v-model.number="limiarForm.limiar_valor" type="number" step="0.01" required />
          </div>
          <button type="submit" class="btn">Atualizar Limiar</button>
        </form>
      </div>

      <h3>Histórico de URLs</h3>
      <ul v-if="monitorHistory.length" class="monitor-history">
        <li v-for="m in monitorHistory" :key="m.id" :class="{ active: m.ativo }">
          <span class="status">{{ m.ativo ? 'Ativo' : 'Arquivado' }}</span>
          <a :href="m.url" target="_blank">{{ m.url }}</a>
          <span class="date">{{ formatDate(m.created_at) }}</span>
        </li>
      </ul>
      <p v-else class="empty">Nenhum monitor configurado.</p>
    </div>
  </div>
  <div v-else-if="loading" class="loading">Carregando...</div>
  <div v-else class="error">{{ error }}</div>
</template>

<script>
import { api } from '../api.js'

export default {
  data() {
    return {
      sku: null,
      loading: true,
      error: null,
      activeTab: 'compra',
      observacoesCompra: [],
      observacoesRastreio: [],
      deltaCompra: null,
      deltaRastreio: null,
      monitor: null,
      monitorHistory: [],
      compraForm: { preco: null, quantidade: null, loja: '', notas: '' },
      monitorForm: { url: '', css_selector: '', regex_pattern: '', limiar_modo: null, limiar_valor: null },
      limiarForm: { limiar_modo: null, limiar_valor: null }
    }
  },
  async mounted() {
    await this.loadSKU()
  },
  methods: {
    async loadSKU() {
      const id = this.$route.params.id
      try {
        this.sku = await api.getSKU(id)
        await Promise.all([
          this.loadObservacoes(),
          this.loadMonitor()
        ])
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    async loadObservacoes() {
      const id = this.sku.id
      const [compra, rastreio] = await Promise.all([
        api.listObservacoes(id, 'compra').catch(() => []),
        api.listObservacoes(id, 'rastreio').catch(() => [])
      ])
      this.observacoesCompra = compra || []
      this.observacoesRastreio = rastreio || []

      this.deltaCompra = await api.getDelta(id, 'compra').catch(() => null)
      this.deltaRastreio = await api.getDelta(id, 'rastreio').catch(() => null)
    },
    async loadMonitor() {
      const id = this.sku.id
      this.monitor = await api.getMonitor(id)
      this.monitorHistory = await api.listMonitorHistory(id).catch(() => [])
      if (this.monitor) {
        this.limiarForm.limiar_modo = this.monitor.limiar_modo
        this.limiarForm.limiar_valor = this.monitor.limiar_valor
      }
    },
    async submitCompra() {
      try {
        const data = {
          sku_id: this.sku.id,
          preco: this.compraForm.preco,
          unidade: this.sku.unidade_padrao,
          loja: this.compraForm.loja,
          fonte: 'compra'
        }
        if (this.compraForm.quantidade) data.quantidade = this.compraForm.quantidade
        if (this.compraForm.notas) data.notas = this.compraForm.notas

        await api.createObservacao(data)
        this.compraForm = { preco: null, quantidade: null, loja: '', notas: '' }
        await this.loadObservacoes()
      } catch (e) {
        alert('Erro: ' + e.message)
      }
    },
    async submitMonitor() {
      try {
        const data = {
          url: this.monitorForm.url
        }
        if (this.monitorForm.css_selector) data.css_selector = this.monitorForm.css_selector
        if (this.monitorForm.regex_pattern) data.regex_pattern = this.monitorForm.regex_pattern
        if (this.monitorForm.limiar_modo) {
          data.limiar_modo = this.monitorForm.limiar_modo
          data.limiar_valor = this.monitorForm.limiar_valor
        }

        await api.setMonitor(this.sku.id, data)
        this.monitorForm = { url: '', css_selector: '', regex_pattern: '', limiar_modo: null, limiar_valor: null }
        await this.loadMonitor()
      } catch (e) {
        alert('Erro: ' + e.message)
      }
    },
    async updateLimiar() {
      try {
        const data = {
          limiar_modo: this.limiarForm.limiar_modo,
          limiar_valor: this.limiarForm.limiar_modo ? this.limiarForm.limiar_valor : null
        }
        await api.updateMonitorLimiar(this.sku.id, data)
        await this.loadMonitor()
      } catch (e) {
        alert('Erro: ' + e.message)
      }
    },
    async deleteSKU() {
      if (!confirm('Excluir este SKU e todas as observações?')) return
      try {
        await api.deleteSKU(this.sku.id)
        this.$router.push('/')
      } catch (e) {
        alert('Erro: ' + e.message)
      }
    },
    formatPrice(val) {
      if (val == null) return '-'
      return parseFloat(val).toFixed(2)
    },
    formatDate(val) {
      return new Date(val).toLocaleString('pt-BR')
    },
    formatDelta(d) {
      if (!d.has_previous) return ''
      const abs = parseFloat(d.delta_absoluto).toFixed(2)
      const pct = parseFloat(d.delta_percentual).toFixed(1)
      const sign = d.delta_absoluto >= 0 ? '+' : ''
      return `${sign}R$ ${abs} (${sign}${pct}%)`
    },
    deltaClass(val) {
      if (val > 0) return 'up'
      if (val < 0) return 'down'
      return ''
    },
    isArchivedUrl(url) {
      return this.monitor && url !== this.monitor.url
    }
  }
}
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.header h1 {
  margin-bottom: 0.5rem;
}

.meta {
  color: #666;
}

.badge {
  display: inline-block;
  background: #3498db;
  color: white;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.btn {
  display: inline-block;
  background: #3498db;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
  font-size: 1rem;
}

.btn:hover {
  background: #2980b9;
}

.btn-danger {
  background: #e74c3c;
}

.btn-danger:hover {
  background: #c0392b;
}

.tabs {
  display: flex;
  gap: 0;
  margin-bottom: 1.5rem;
  border-bottom: 2px solid #ddd;
}

.tabs button {
  padding: 0.75rem 1.5rem;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 1rem;
  color: #666;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
}

.tabs button.active {
  color: #3498db;
  border-bottom-color: #3498db;
}

.tab-content {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.tab-content h3 {
  margin: 1.5rem 0 1rem;
}

.tab-content h3:first-child {
  margin-top: 0;
}

.delta-card {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.delta-card h3 {
  margin: 0 0 0.5rem !important;
  font-size: 0.9rem;
  color: #666;
}

.price {
  font-size: 2rem;
  font-weight: bold;
  color: #2c3e50;
}

.delta {
  font-size: 1.1rem;
  margin-top: 0.25rem;
}

.delta.up {
  color: #e74c3c;
}

.delta.down {
  color: #27ae60;
}

.form {
  max-width: 400px;
}

.form-row {
  margin-bottom: 1rem;
}

.form-row label {
  display: block;
  margin-bottom: 0.25rem;
  font-weight: 500;
}

.form-row input,
.form-row select,
.form-row textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.form-row textarea {
  min-height: 80px;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.table th {
  background: #f8f9fa;
  font-weight: 600;
}

.table tr.archived {
  opacity: 0.5;
}

.url {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.monitor-info {
  background: #e8f4fd;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.monitor-info h3 {
  margin: 0 0 0.5rem !important;
}

.monitor-info a {
  color: #3498db;
  word-break: break-all;
}

.monitor-history {
  list-style: none;
}

.monitor-history li {
  padding: 0.75rem;
  border-bottom: 1px solid #eee;
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.monitor-history li.active {
  background: #e8f4fd;
}

.monitor-history .status {
  font-size: 0.8rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: #95a5a6;
  color: white;
}

.monitor-history li.active .status {
  background: #27ae60;
}

.monitor-history a {
  color: #3498db;
  flex: 1;
  word-break: break-all;
}

.monitor-history .date {
  color: #666;
  font-size: 0.9rem;
}

.empty {
  color: #666;
  font-style: italic;
}

.loading, .error {
  text-align: center;
  padding: 2rem;
}

.error {
  color: #e74c3c;
}
</style>
