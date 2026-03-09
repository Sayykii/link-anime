<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useLibraryStore } from '@/stores/library'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import {
  LayoutDashboard,
  Library,
  Link,
  Download,
  Rss,
  History,
  Settings,
  Tv,
  Film,
} from 'lucide-vue-next'

const model = defineModel<boolean>({ default: false })

const router = useRouter()
const library = useLibraryStore()

const navigationItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Library', path: '/library', icon: Library },
  { name: 'Link Wizard', path: '/link', icon: Link },
  { name: 'Downloads', path: '/downloads', icon: Download },
  { name: 'RSS Watch', path: '/rss', icon: Rss },
  { name: 'History', path: '/history', icon: History },
  { name: 'Settings', path: '/settings', icon: Settings },
]

const shows = computed(() => library.shows ?? [])
const movies = computed(() => library.movies ?? [])

function navigate(path: string) {
  model.value = false
  router.push(path)
}

function navigateToShow(name: string) {
  model.value = false
  router.push('/library')
}

function navigateToMovie(name: string) {
  model.value = false
  router.push('/library')
}
</script>

<template>
  <CommandDialog v-model:open="model" title="Command Palette" description="Search commands, shows, movies...">
    <CommandInput placeholder="Type a command or search..." />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>

      <CommandGroup heading="Navigation">
        <CommandItem
          v-for="item in navigationItems"
          :key="item.path"
          :value="'nav-' + item.name"
          @select="navigate(item.path)"
          class="gap-2"
        >
          <component :is="item.icon" class="h-4 w-4 text-muted-foreground" />
          <span>{{ item.name }}</span>
        </CommandItem>
      </CommandGroup>

      <template v-if="shows.length">
        <CommandSeparator />
        <CommandGroup heading="Shows">
          <CommandItem
            v-for="show in shows"
            :key="show.path"
            :value="'show-' + show.name"
            @select="navigateToShow(show.name)"
            class="gap-2"
          >
            <Tv class="h-4 w-4 text-muted-foreground" />
            <span>{{ show.name }}</span>
            <span class="ml-auto text-xs text-muted-foreground">
              {{ show.seasons.length }} season{{ show.seasons.length !== 1 ? 's' : '' }}
            </span>
          </CommandItem>
        </CommandGroup>
      </template>

      <template v-if="movies.length">
        <CommandSeparator />
        <CommandGroup heading="Movies">
          <CommandItem
            v-for="movie in movies"
            :key="movie.path"
            :value="'movie-' + movie.name"
            @select="navigateToMovie(movie.name)"
            class="gap-2"
          >
            <Film class="h-4 w-4 text-muted-foreground" />
            <span>{{ movie.name }}</span>
          </CommandItem>
        </CommandGroup>
      </template>
    </CommandList>
  </CommandDialog>
</template>
