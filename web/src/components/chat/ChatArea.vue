<template>
  <div class="chat-area">
    <!-- Welcome Screen -->
    <div v-if="messages.length === 0" class="welcome-screen">
      <div class="logo">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"></circle>
          <path d="M12 16v-4"></path>
          <path d="M12 8h.01"></path>
        </svg>
      </div>
      <h1>How can I help you today?</h1>
    </div>

    <!-- Message List -->
    <div v-else class="message-list" ref="messageListRef">
      <MessageBubble
        v-for="(msg, index) in messages"
        :key="index"
        :role="msg.role"
        :content="msg.content"
        :is-streaming="msg.isStreaming"
      />
    </div>

    <!-- Input Area -->
    <InputArea
      :disabled="isLoading"
      :placeholder="isLoading ? 'Generating...' : 'Message Coca AI...'"
      @send="handleSend"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import MessageBubble from './MessageBubble.vue'
import InputArea from './InputArea.vue'

export interface DisplayMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  isStreaming?: boolean
}

const props = defineProps<{
  messages: DisplayMessage[]
  isLoading: boolean
}>()

const emit = defineEmits<{
  (e: 'send', message: string): void
}>()

const messageListRef = ref<HTMLElement | null>(null)

const handleSend = (message: string) => {
  emit('send', message)
}

// Auto scroll to bottom when messages change
watch(
  () => props.messages,
  () => {
    nextTick(() => {
      if (messageListRef.value) {
        messageListRef.value.scrollTop = messageListRef.value.scrollHeight
      }
    })
  },
  { deep: true }
)
</script>

<style scoped>
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #212121;
  overflow: hidden;
}

.welcome-screen {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #ececec;
}

.logo {
  color: #10a37f;
  margin-bottom: 24px;
}

.welcome-screen h1 {
  font-size: 28px;
  font-weight: 600;
  margin: 0;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px;
}

/* Scrollbar styling */
.message-list::-webkit-scrollbar {
  width: 8px;
}

.message-list::-webkit-scrollbar-track {
  background: transparent;
}

.message-list::-webkit-scrollbar-thumb {
  background: #3f3f3f;
  border-radius: 4px;
}

.message-list::-webkit-scrollbar-thumb:hover {
  background: #5f5f5f;
}
</style>
