<template>
  <div class="app">
    <header>
      <nav>
        <router-link to="/" class="brand">Spendfy</router-link>
        <router-link to="/skus/new" class="btn-new">+ Novo SKU</router-link>
      </nav>
    </header>
    <main>
      <router-view />
    </main>
  </div>
</template>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  background: #f5f5f5;
  color: #333;
  line-height: 1.6;
}

.app {
  min-height: 100vh;
}

header {
  background: #2c3e50;
  padding: 1rem 2rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

nav {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.brand {
  color: white;
  text-decoration: none;
  font-size: 1.5rem;
  font-weight: bold;
}

.btn-new {
  background: #27ae60;
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  text-decoration: none;
  transition: background 0.2s;
}

.btn-new:hover {
  background: #219a52;
}

main {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 0 1rem;
}
</style>
