import { ref, watch, type Ref } from 'vue'
import { useTransition, TransitionPresets } from '@vueuse/core'

export function useCountUp(
  source: Ref<number>,
  options: { duration?: number; delay?: number } = {},
) {
  const { duration = 600, delay = 0 } = options
  const outputSource = ref(0)

  const animated = useTransition(outputSource, {
    duration,
    transition: TransitionPresets.easeOutCubic,
    delay,
  })

  watch(
    source,
    (val) => {
      outputSource.value = val
    },
    { immediate: true },
  )

  return animated
}
