<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goHome() {
  router.push('/')
}
</script>

<template>
  <div class="not-found">
    <div class="content">
      <h1>404</h1>
      <h2>页面未找到</h2>
      <p>抱歉，您访问的页面不存在</p>
      <button class="btn btn-primary" @click="goHome">返回首页</button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.not-found {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100vh;
  background-color: $bg-color-dark;
}

.content {
  text-align: center;
}

h1 {
  font-size: 120px;
  font-weight: 700;
  color: $primary-color;
  margin-bottom: $spacing-md;
  line-height: 1;
}

h2 {
  font-size: $font-size-xl;
  font-weight: 500;
  color: $text-color-primary;
  margin-bottom: $spacing-md;
}

p {
  font-size: $font-size-base;
  color: $text-color-secondary;
  margin-bottom: $spacing-lg;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-lg;
  background-color: $bg-color-light;
  border: 1px solid $border-color;
  border-radius: $border-radius-base;
  color: $text-color-primary;
  font-size: $font-size-base;
  cursor: pointer;
  transition: all $transition-duration $transition-timing;

  &.btn-primary {
    background-color: $primary-color;
    border-color: $primary-color;
    color: white;

    &:hover {
      background-color: lighten($primary-color, 10%);
    }
  }
}
</style>
