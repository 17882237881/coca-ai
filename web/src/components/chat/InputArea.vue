<template>
  <div class="input-area">
    <div class="input-container">
      <textarea
        ref="textareaRef"
        v-model="message"
        :placeholder="placeholder"
        :disabled="disabled"
        @keydown="handleKeydown"
        @input="autoResize"
        rows="1"
      ></textarea>
      <button
        class="send-btn"
        :disabled="!canSend"
        @click="handleSend"
      >
        <svg v-if="!disabled" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
        </svg>
        <div v-else class="loading-spinner"></div>
      </button>
    </div>
    <p class="disclaimer">Coca AI may produce inaccurate information. Consider checking important info.</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'

const props = defineProps<{
  disabled?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'send', message: string): void
}>()

const message = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const canSend = computed(() => message.value.trim().length > 0 && !props.disabled)

const handleSend = () => {
  if (!canSend.value) return
  emit('send', message.value.trim())
  message.value = ''
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.style.height = 'auto'
    }
  })
}

const handleKeydown = (e: KeyboardEvent) => {
  // Enter to send, Shift+Enter for new line
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

const autoResize = () => {
  const textarea = textareaRef.value
  if (textarea) {
    textarea.style.height = 'auto'
    textarea.style.height = Math.min(textarea.scrollHeight, 200) + 'px'
  }
}
</script>

<style scoped>
.input-area {
  padding: 16px 24px 24px;
  background: #212121;
}

.input-container {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  max-width: 800px;
  margin: 0 auto;
  background: #2f2f2f;
  border-radius: 16px;
  padding: 12px 16px;
  border: 1px solid #3f3f3f;
  transition: border-color 0.2s;
}

.input-container:focus-within {
  border-color: #10a37f;
}

textarea {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #ececec;
  font-size: 15px;
  line-height: 1.5;
  resize: none;
  max-height: 200px;
  font-family: inherit;
}

textarea::placeholder {
  color: #8e8e8e;
}

textarea:disabled {
  opacity: 0.6;
}

.send-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: #10a37f;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s, opacity 0.2s;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
  background: #0d8a6a;
}

.send-btn:disabled {
  background: #3f3f3f;
  color: #8e8e8e;
  cursor: not-allowed;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #8e8e8e;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.disclaimer {
  text-align: center;
  font-size: 12px;
  color: #8e8e8e;
  margin-top: 12px;
}
</style>
