<template>
  <div class="sku-form">
    <h1>{{ isEdit ? 'Editar SKU' : 'Novo SKU' }}</h1>

    <form @submit.prevent="submit" class="form">
      <div class="form-row">
        <label>Nome *</label>
        <input v-model="form.nome" type="text" required placeholder="Ex: Vasenol Clinical 200ml" />
      </div>

      <div class="form-row">
        <label>Marca</label>
        <input v-model="form.marca" type="text" placeholder="Ex: Vasenol" />
      </div>

      <div class="form-row">
        <label>Tamanho ou Variante</label>
        <input v-model="form.tamanho_ou_variante" type="text" placeholder="Ex: 200ml" />
      </div>

      <div class="form-row">
        <label>Unidade Padrão *</label>
        <select v-model="form.unidade_padrao" required>
          <option value="un">Unidade (un)</option>
          <option value="kg">Quilograma (kg)</option>
          <option value="g">Grama (g)</option>
          <option value="l">Litro (l)</option>
          <option value="ml">Mililitro (ml)</option>
        </select>
      </div>

      <div class="form-row" v-if="!isEdit">
        <label>Chave de Identidade *</label>
        <input v-model="form.chave_identidade" type="text" required 
               placeholder="Ex: vasenol-clinical-200ml" />
        <small>Identificador único. Sugestão: marca-nome-tamanho em minúsculas.</small>
      </div>

      <div class="actions">
        <button type="submit" class="btn" :disabled="submitting">
          {{ submitting ? 'Salvando...' : 'Salvar' }}
        </button>
        <router-link to="/" class="btn btn-secondary">Cancelar</router-link>
      </div>

      <p v-if="error" class="error">{{ error }}</p>
    </form>
  </div>
</template>

<script>
import { api } from '../api.js'

export default {
  data() {
    return {
      form: {
        nome: '',
        marca: '',
        tamanho_ou_variante: '',
        unidade_padrao: 'un',
        chave_identidade: ''
      },
      submitting: false,
      error: null
    }
  },
  computed: {
    isEdit() {
      return !!this.$route.params.id
    }
  },
  async mounted() {
    if (this.isEdit) {
      try {
        const sku = await api.getSKU(this.$route.params.id)
        this.form = {
          nome: sku.nome,
          marca: sku.marca || '',
          tamanho_ou_variante: sku.tamanho_ou_variante || '',
          unidade_padrao: sku.unidade_padrao,
          chave_identidade: sku.chave_identidade
        }
      } catch (e) {
        this.error = e.message
      }
    }
  },
  methods: {
    async submit() {
      this.submitting = true
      this.error = null

      try {
        const data = {
          nome: this.form.nome,
          unidade_padrao: this.form.unidade_padrao
        }
        if (this.form.marca) data.marca = this.form.marca
        if (this.form.tamanho_ou_variante) data.tamanho_ou_variante = this.form.tamanho_ou_variante

        if (this.isEdit) {
          await api.updateSKU(this.$route.params.id, data)
          this.$router.push(`/skus/${this.$route.params.id}`)
        } else {
          data.chave_identidade = this.form.chave_identidade
          const sku = await api.createSKU(data)
          this.$router.push(`/skus/${sku.id}`)
        }
      } catch (e) {
        this.error = e.message
      } finally {
        this.submitting = false
      }
    }
  }
}
</script>

<style scoped>
.sku-form {
  max-width: 500px;
}

.sku-form h1 {
  margin-bottom: 1.5rem;
}

.form {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
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
.form-row select {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.form-row small {
  display: block;
  margin-top: 0.25rem;
  color: #666;
  font-size: 0.85rem;
}

.actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1.5rem;
}

.btn {
  display: inline-block;
  background: #3498db;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
  font-size: 1rem;
}

.btn:hover {
  background: #2980b9;
}

.btn:disabled {
  background: #95a5a6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #95a5a6;
}

.btn-secondary:hover {
  background: #7f8c8d;
}

.error {
  color: #e74c3c;
  margin-top: 1rem;
}
</style>
