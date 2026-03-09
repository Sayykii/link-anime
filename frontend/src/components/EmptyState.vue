<script setup lang="ts">
import type { Component } from 'vue'
import { Button } from '@/components/ui/button'

defineProps<{
  icon: Component
  heading: string
  description?: string
  actionLabel?: string
  actionVariant?: 'default' | 'outline' | 'secondary'
}>()

defineEmits<{
  action: []
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 text-center">
    <div class="animate-float mb-4 text-muted-foreground/40">
      <component :is="icon" class="h-12 w-12" />
    </div>
    <p class="font-medium text-foreground">{{ heading }}</p>
    <p v-if="description" class="mt-1 text-sm text-muted-foreground max-w-sm">{{ description }}</p>
    <Button
      v-if="actionLabel"
      :variant="actionVariant ?? 'default'"
      class="mt-4"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </Button>
  </div>
</template>
