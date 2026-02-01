<template>
  <div class="message-bubble" :class="{ 'user-message': isUser, 'assistant-message': !isUser }">
    <div class="avatar">
      <div v-if="isUser" class="user-avatar">U</div>
      <div v-else class="ai-avatar">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <path d="M12 16v-4"></path>
          <path d="M12 8h.01"></path>
        </svg>
      </div>
    </div>
    <div class="content">
      <div class="role-name">{{ isUser ? 'You' : 'Coca AI' }}</div>
      <div class="message-text" v-html="renderedContent"></div>
      <div v-if="isStreaming" class="cursor">▍</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import hljs from 'highlight.js'

const props = defineProps<{
  role: 'user' | 'assistant' | 'system'
  content: string
  isStreaming?: boolean
}>()

const isUser = computed(() => props.role === 'user')

// Configure marked for syntax highlighting
marked.setOptions({
  highlight: function(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value
      } catch (e) {
        console.error(e)
      }
    }
    return hljs.highlightAuto(code).value
  },
  breaks: true,
})

const renderedContent = computed(() => {
  if (isUser.value) {
    // User messages: preserve line breaks but don't render markdown
    return props.content.replace(/\n/g, '<br>')
  }
  // AI messages: render markdown
  return marked.parse(props.content)
})
</script>

<style scoped>
.message-bubble {
  display: flex;
  gap: 16px;
  padding: 24px 0;
  max-width: 800px;
  margin: 0 auto;
}

.avatar {
  flex-shrink: 0;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  background: #5436DA;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.ai-avatar {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  background: #10a37f;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

.content {
  flex: 1;
  min-width: 0;
}

.role-name {
  font-size: 14px;
  font-weight: 600;
  color: #ececec;
  margin-bottom: 8px;
}

.message-text {
  color: #d1d5db;
  font-size: 15px;
  line-height: 1.7;
  word-wrap: break-word;
}

.message-text :deep(p) {
  margin: 0 0 12px 0;
}

.message-text :deep(p:last-child) {
  margin-bottom: 0;
}

.message-text :deep(pre) {
  background: #1e1e1e;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
  margin: 12px 0;
}

.message-text :deep(code) {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
}

.message-text :deep(code:not(pre code)) {
  background: #3f3f3f;
  padding: 2px 6px;
  border-radius: 4px;
}

.message-text :deep(ul),
.message-text :deep(ol) {
  padding-left: 24px;
  margin: 12px 0;
}

.message-text :deep(li) {
  margin: 4px 0;
}

.message-text :deep(a) {
  color: #10a37f;
  text-decoration: none;
}

.message-text :deep(a:hover) {
  text-decoration: underline;
}

.message-text :deep(blockquote) {
  border-left: 3px solid #10a37f;
  padding-left: 16px;
  margin: 12px 0;
  color: #9ca3af;
}

.cursor {
  display: inline-block;
  animation: blink 1s infinite;
  color: #10a37f;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
</style>
