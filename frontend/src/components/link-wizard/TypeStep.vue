<script setup lang="ts">
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

defineProps<{
  sourceName: string
}>()

defineEmits<{
  'select-type': [type: 'series' | 'movie']
  back: []
}>()
</script>

<template>
  <Card glass>
    <CardHeader>
      <CardTitle>Select Type</CardTitle>
      <CardDescription>
        Is "{{ sourceName }}" a series or a movie?
      </CardDescription>
    </CardHeader>
    <CardContent class="flex gap-4">
      <Button
        size="lg"
        variant="outline"
        class="flex-1 gap-2"
        @click="$emit('select-type', 'series')"
      >
        Series
      </Button>
      <Button
        size="lg"
        variant="outline"
        class="flex-1 gap-2"
        @click="$emit('select-type', 'movie')"
      >
        Movie
      </Button>
    </CardContent>
    <CardContent class="pt-0">
      <Button variant="ghost" size="sm" @click="$emit('back')">Back</Button>
    </CardContent>
  </Card>
</template>
