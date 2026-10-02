<template>
  <section>
    <header class="page-header">
      <p class="eyebrow">设备运营</p>
      <h1>设备管理</h1>
      <p class="page-description">
        设备列表将在业务服务接入后显示。这里用于查看绑定关系、在线状态和版本信息。
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
  margin: 12px 0 0;
  color: var(--sprout-text-muted);
}
</style>
