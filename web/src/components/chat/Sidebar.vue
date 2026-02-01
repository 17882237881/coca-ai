<template>
  <aside class="sidebar">
    <!-- New Chat Button -->
    <button class="new-chat-btn" @click="$emit('newChat')">
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19"></line>
        <line x1="5" y1="12" x2="19" y2="12"></line>
      </svg>
      New chat
    </button>

    <!-- Session List -->
    <div class="session-list">
      <div
        v-for="session in sessions"
        :key="session.session_id"
        class="session-item"
        :class="{ active: session.session_id === currentSessionId }"
        @click="$emit('selectSession', session.session_id)"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
        <span class="session-title">{{ session.title }}</span>
        <button
          class="delete-btn"
          @click.stop="$emit('deleteSession', session.session_id)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- User Menu -->
    <div class="user-menu">
      <button class="logout-btn" @click="$emit('logout')">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
          <polyline points="16 17 21 12 16 7"></polyline>
          <line x1="21" y1="12" x2="9" y2="12"></line>
        </svg>
        Logout
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import type { Session } from '../api/chat'

defineProps<{
  sessions: Session[]
  currentSessionId: number | null
}>()

defineEmits<{
  (e: 'newChat'): void
  (e: 'selectSession', sessionId: number): void
  (e: 'deleteSession', sessionId: number): void
  (e: 'logout'): void
}>()
</script>

<style scoped>
.sidebar {
  width: 260px;
  height: 100vh;
  background: #171717;
  display: flex;
  flex-direction: column;
  padding: 8px;
  border-right: 1px solid #2f2f2f;
}

.new-chat-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  background: transparent;
  border: 1px solid #3f3f3f;
  border-radius: 8px;
  color: #ececec;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}

.new-chat-btn:hover {
  background: #2f2f2f;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  margin-top: 16px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  color: #ececec;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
  position: relative;
}

.session-item:hover {
  background: #2f2f2f;
}

.session-item.active {
  background: #343541;
}

.session-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.delete-btn {
  display: none;
  padding: 4px;
  background: transparent;
  border: none;
  color: #8e8e8e;
  cursor: pointer;
  border-radius: 4px;
}

.delete-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.session-item:hover .delete-btn {
  display: block;
}

.user-menu {
  border-top: 1px solid #2f2f2f;
  padding-top: 8px;
  margin-top: 8px;
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: #8e8e8e;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}

.logout-btn:hover {
  background: #2f2f2f;
  color: #ececec;
}
</style>
