<template>
  <div class="sku-list">
    <h1>SKUs</h1>
    
    <div v-if="loading" class="loading">Carregando...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="skus.length === 0" class="empty">
      <p>Nenhum SKU cadastrado.</p>
      <router-link to="/skus/new" class="btn">Criar primeiro SKU</router-link>
    </div>
    <div v-else class="grid">
      <div v-for="sku in skus" :key="sku.id" class="card">
        <router-link :to="`/skus/${sku.id}`" class="card-link">
          <h3>{{ sku.nome }}</h3>
          <p v-if="sku.marca" class="meta">{{ sku.marca }}</p>
          <p v-if="sku.tamanho_ou_variante" class="meta">{{ sku.tamanho_ou_variante }}</p>
          <span class="badge">{{ sku.unidade_padrao }}</span>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
import { api } from '../api.js'

export default {
  data() {
    return {
      skus: [],
      loading: true,
      error: null
    }
  },
  async mounted() {
    try {
      const data = await api.listSKUs()
      this.skus = data || []
    } catch (e) {
      this.error = e.message
    } finally {
      this.loading = false
    }
  }
}
</script>

<style scoped>
.sku-list h1 {
  margin-bottom: 1.5rem;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0,0,0,0.15);
}

.card-link {
  display: block;
  padding: 1.5rem;
  text-decoration: none;
  color: inherit;
}

.card h3 {
  margin-bottom: 0.5rem;
  color: #2c3e50;
}

.meta {
  color: #666;
  font-size: 0.9rem;
  margin-bottom: 0.25rem;
}

.badge {
  display: inline-block;
  background: #3498db;
  color: white;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  margin-top: 0.5rem;
}

.loading, .error, .empty {
  text-align: center;
  padding: 2rem;
}

.error {
  color: #e74c3c;
}

.btn {
  display: inline-block;
  background: #3498db;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  text-decoration: none;
  margin-top: 1rem;
}
</style>
