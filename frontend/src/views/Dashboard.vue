<template>
  <div class="dashboard">
    <h1>Dashboard de Produtos Rastreados</h1>
    
    <div v-if="loading" class="loading">Carregando...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="trackedSkus.length === 0" class="empty">
      <p>Nenhum produto rastreado.</p>
      <p>Adicione observações ou monitores aos SKUs para vê-los aqui.</p>
      <router-link to="/skus" class="btn">Ver SKUs</router-link>
    </div>
    <div v-else class="cards">
      <div v-for="item in trackedSkus" :key="item.sku.id" class="card">
        <router-link :to="`/skus/${item.sku.id}`" class="card-header">
          <h3>{{ item.sku.nome }}</h3>
          <div class="meta">
            <span v-if="item.sku.marca">{{ item.sku.marca }}</span>
            <span v-if="item.sku.tamanho_ou_variante">{{ item.sku.tamanho_ou_variante }}</span>
          </div>
          <div class="badges">
            <span class="badge unit">{{ item.sku.unidade_padrao }}</span>
            <span v-if="item.has_monitor" class="badge monitor">Monitor ativo</span>
          </div>
        </router-link>
        
        <div class="card-body">
          <div class="last-compra">
            <span class="label">Último valor pago:</span>
            <span class="value" v-if="item.last_compra">
              R$ {{ formatPrice(item.last_compra) }}
              <span class="date">({{ formatDate(item.last_compra_data) }})</span>
            </span>
            <span class="value none" v-else>—</span>
          </div>
          
          <div class="chart-container" v-if="hasSeries(item)">
            <Line :data="getChartData(item)" :options="chartOptions" />
          </div>
          <div class="no-chart" v-else>
            <span>Sem dados suficientes para gráfico</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { api } from '../api.js'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

export default {
  components: { Line },
  data() {
    return {
      trackedSkus: [],
      loading: true,
      error: null,
      chartOptions: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: 'index',
          intersect: false
        },
        plugins: {
          legend: {
            position: 'bottom',
            labels: { usePointStyle: true, boxWidth: 8 }
          },
          tooltip: {
            callbacks: {
              label: (ctx) => `${ctx.dataset.label}: R$ ${parseFloat(ctx.raw).toFixed(2)}`
            }
          }
        },
        scales: {
          x: {
            grid: { display: false },
            ticks: { maxTicksLimit: 6 }
          },
          y: {
            beginAtZero: false,
            ticks: {
              callback: (val) => `R$ ${val}`
            }
          }
        }
      }
    }
  },
  async mounted() {
    try {
      const data = await api.listTrackedSKUs()
      this.trackedSkus = data || []
    } catch (e) {
      this.error = e.message
    } finally {
      this.loading = false
    }
  },
  methods: {
    formatPrice(val) {
      if (val == null) return '-'
      return parseFloat(val).toFixed(2)
    },
    formatDate(val) {
      if (!val) return ''
      return new Date(val).toLocaleDateString('pt-BR')
    },
    hasSeries(item) {
      const byUrl = item.series_rastreio_by_url || []
      const rastPts = byUrl.reduce((n, s) => n + ((s.points && s.points.length) || 0), 0)
      return (item.series_compra && item.series_compra.length > 0) || rastPts > 0
    },
    hostLabel(url) {
      try { return new URL(url).hostname.replace(/^www\./, '') } catch (e) { return url }
    },
    getChartData(item) {
      const compra = item.series_compra || []
      const byUrl = item.series_rastreio_by_url || []
      const colors = ['#3498db', '#9b59b6', '#e67e22', '#1abc9c', '#e74c3c', '#34495e']
      const toMs = (val) => new Date(val).getTime()
      const labelFor = (val) =>
        new Date(val).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: '2-digit' })

      const allPts = [...compra]
      byUrl.forEach(s => (s.points || []).forEach(pt => allPts.push(pt)))
      const axisMap = new Map()
      for (const pt of allPts) {
        const ms = toMs(pt.data)
        if (!axisMap.has(ms)) axisMap.set(ms, labelFor(pt.data))
      }
      const axis = [...axisMap.entries()].sort((a, b) => a[0] - b[0]).map(([ms, label]) => ({ t: ms, label }))

      const datasets = []
      if (compra.length > 0) {
        const compraMap = new Map()
        compra.forEach(pt => compraMap.set(toMs(pt.data), parseFloat(pt.preco)))
        datasets.push({
          label: 'Compra',
          data: axis.map(a => (compraMap.has(a.t) ? compraMap.get(a.t) : null)),
          borderColor: '#27ae60',
          backgroundColor: 'rgba(39, 174, 96, 0.1)',
          tension: 0.3,
          pointRadius: 4,
          pointHoverRadius: 6,
          spanGaps: true
        })
      }
      byUrl.forEach((s, i) => {
        const map = new Map()
        ;(s.points || []).forEach(pt => map.set(toMs(pt.data), parseFloat(pt.preco)))
        const color = colors[i % colors.length]
        datasets.push({
          label: 'Rastreio · ' + this.hostLabel(s.url),
          data: axis.map(a => (map.has(a.t) ? map.get(a.t) : null)),
          borderColor: color,
          backgroundColor: color + '1a',
          tension: 0.3,
          pointRadius: 4,
          pointHoverRadius: 6,
          spanGaps: true
        })
      })
      return { labels: axis.map(a => a.label), datasets }
    }
  }
}
</script>

<style scoped>
.dashboard h1 {
  margin-bottom: 1.5rem;
}

.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 1.5rem;
}

.card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
}

.card-header {
  display: block;
  padding: 1.25rem;
  text-decoration: none;
  color: inherit;
  border-bottom: 1px solid #eee;
  transition: background 0.15s;
}

.card-header:hover {
  background: #f8f9fa;
}

.card-header h3 {
  margin: 0 0 0.5rem;
  color: #2c3e50;
  font-size: 1.15rem;
}

.meta {
  color: #666;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
}

.meta span:not(:last-child)::after {
  content: ' · ';
}

.badges {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.badge.unit {
  background: #3498db;
  color: white;
}

.badge.monitor {
  background: #27ae60;
  color: white;
}

.card-body {
  padding: 1.25rem;
}

.last-compra {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.last-compra .label {
  font-size: 0.85rem;
  color: #666;
}

.last-compra .value {
  font-size: 1.25rem;
  font-weight: 600;
  color: #2c3e50;
}

.last-compra .value.none {
  color: #999;
}

.last-compra .date {
  font-size: 0.8rem;
  font-weight: normal;
  color: #999;
}

.chart-container {
  height: 180px;
  position: relative;
}

.no-chart {
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f9fa;
  border-radius: 8px;
  color: #999;
  font-size: 0.9rem;
}

.loading, .error, .empty {
  text-align: center;
  padding: 3rem 1rem;
  background: white;
  border-radius: 12px;
}

.error {
  color: #e74c3c;
}

.empty p {
  margin-bottom: 0.5rem;
  color: #666;
}

.btn {
  display: inline-block;
  background: #3498db;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 6px;
  text-decoration: none;
  margin-top: 1rem;
  transition: background 0.15s;
}

.btn:hover {
  background: #2980b9;
}
</style>
