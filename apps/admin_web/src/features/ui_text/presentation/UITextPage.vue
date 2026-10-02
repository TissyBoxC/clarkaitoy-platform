<template>
  <section>
    <header class="page-header">
      <p class="eyebrow">界面文案</p>
      <h1>界面文案管理</h1>
      <p class="page-description">
        文案发布后会按应用、设备和语言分批生效。未发布的修改不会被用户看到。
      </p>
    </header>
  </section>
</template>

<style scoped>
.page-header {
  padding: 24px;
  border: 1px solid var(--sprout-outline);
  border-radius: var(--sprout-radius-card);
  background: #ffffff;
  box-shadow: 0 8px 24px rgb(194 91 128 / 5%);
}

.eyebrow {
  margin: 0 0 6px;
  color: var(--sprout-pink-strong);
  font-size: 13px;
  font-weight: 700;
}

h1 {
  margin: 0;
  color: var(--sprout-text);
  font-size: 28px;
}

.page-description {
  max-width: 680px;
  margin: 12px 0 0;
  color: var(--sprout-text-muted);
}
</style>
