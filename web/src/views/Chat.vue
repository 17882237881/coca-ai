<template>
  <div class="chat-container">
    <!-- Sidebar -->
    <Sidebar
      :sessions="sessions"
      :current-session-id="currentSessionId"
      @new-chat="handleNewChat"
      @select-session="handleSelectSession"
      @delete-session="handleDeleteSession"
      @logout="handleLogout"
    />

    <!-- Chat Area -->
    <ChatArea
      :messages="messages"
      :is-loading="isLoading"
      @send="handleSendMessage"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Sidebar from '../components/chat/Sidebar.vue'
import ChatArea from '../components/chat/ChatArea.vue'
import type { DisplayMessage } from '../components/chat/ChatArea.vue'
import {
  getSessions,
  createSession,
  deleteSession,
  getMessages,
  sendMessageStream,
  type Session,
} from '../api/chat'
import request from '../api/request'

const router = useRouter()

// State
const sessions = ref<Session[]>([])
const currentSessionId = ref<number | null>(null)
const messages = ref<DisplayMessage[]>([])
const isLoading = ref(false)

// Load sessions on mount
onMounted(async () => {
  await loadSessions()
})

const loadSessions = async () => {
  try {
    sessions.value = await getSessions()
  } catch (e) {
    console.error('Failed to load sessions', e)
  }
}

const loadMessages = async (sessionId: number) => {
  try {
    const msgs = await getMessages(sessionId)
    messages.value = msgs.map(m => ({
      role: m.role as 'user' | 'assistant',
      content: m.content,
    }))
  } catch (e) {
    console.error('Failed to load messages', e)
    messages.value = []
  }
}

const handleNewChat = async () => {
  try {
    const session = await createSession()
    sessions.value.unshift(session)
    currentSessionId.value = session.session_id
    messages.value = []
  } catch (e) {
    console.error('Failed to create session', e)
  }
}

const handleSelectSession = async (sessionId: number) => {
  currentSessionId.value = sessionId
  await loadMessages(sessionId)
}

const handleDeleteSession = async (sessionId: number) => {
  try {
    await deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (currentSessionId.value === sessionId) {
      currentSessionId.value = null
      messages.value = []
    }
  } catch (e) {
    console.error('Failed to delete session', e)
  }
}

const handleLogout = async () => {
  try {
    await request.post('/users/logout')
  } catch (e) {
    console.error('Logout failed', e)
  } finally {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    router.push('/login')
  }
}

const handleSendMessage = async (content: string) => {
  // Create session if needed
  if (!currentSessionId.value) {
    try {
      const session = await createSession()
      sessions.value.unshift(session)
      currentSessionId.value = session.session_id
    } catch (e) {
      console.error('Failed to create session', e)
      return
    }
  }

  // Add user message
  messages.value.push({
    role: 'user',
    content: content,
  })

  // Add empty assistant message for streaming
  const assistantIndex = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '',
    isStreaming: true,
  })

  isLoading.value = true

  // Send message with SSE streaming
  sendMessageStream(
    currentSessionId.value,
    content,
    {
      onMessage: (delta) => {
        messages.value[assistantIndex].content += delta
      },
      onDone: () => {
        messages.value[assistantIndex].isStreaming = false
        isLoading.value = false
        // Reload sessions to update titles
        loadSessions()
      },
      onError: (error) => {
        messages.value[assistantIndex].content = `Error: ${error}`
        messages.value[assistantIndex].isStreaming = false
        isLoading.value = false
      },
    }
  )
}
</script>

<style scoped>
.chat-container {
  display: flex;
  height: 100vh;
  background: #212121;
}
</style>

<style>
/* Import highlight.js theme */
@import 'highlight.js/styles/github-dark.css';

/* Global styles */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: 'Söhne', 'Segoe UI', 'Helvetica Neue', sans-serif;
  background: #212121;
  color: #ececec;
}
</style>
