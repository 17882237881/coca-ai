import request from './request'

// ==================== Types ====================

export interface Session {
    session_id: number
    title: string
    updated_at: string
}

export interface Message {
    id: number
    role: 'user' | 'assistant' | 'system'
    content: string
    created_at: string
}

// ==================== Session APIs ====================

export async function createSession(): Promise<Session> {
    const res = await request.post<{ data: Session }>('/chat/sessions')
    return res.data.data
}

export async function getSessions(): Promise<Session[]> {
    const res = await request.get<{ data: Session[] }>('/chat/sessions')
    return res.data.data || []
}

export async function deleteSession(sessionId: number): Promise<void> {
    await request.delete(`/chat/sessions/${sessionId}`)
}

// ==================== Message APIs ====================

export async function getMessages(sessionId: number): Promise<Message[]> {
    const res = await request.get<{ data: Message[] }>(`/chat/sessions/${sessionId}/messages`)
    return res.data.data || []
}

// ==================== SSE Streaming ====================

export interface SSECallbacks {
    onMessage: (delta: string) => void
    onDone: (messageId: number, content: string) => void
    onError: (error: string) => void
}

export function sendMessageStream(
    sessionId: number,
    content: string,
    callbacks: SSECallbacks
): () => void {
    const token = localStorage.getItem('access_token')
    const url = `http://localhost:8080/chat/sessions/${sessionId}/messages`

    // Using fetch with streaming for SSE with POST
    const controller = new AbortController()

    fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ content }),
        signal: controller.signal,
    })
        .then(async (response) => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`)
            }

            const reader = response.body?.getReader()
            if (!reader) {
                throw new Error('No reader available')
            }

            const decoder = new TextDecoder()
            let buffer = ''

            while (true) {
                const { done, value } = await reader.read()
                if (done) break

                buffer += decoder.decode(value, { stream: true })

                // Parse SSE events
                const lines = buffer.split('\n')
                buffer = lines.pop() || ''

                for (const line of lines) {
                    if (line.startsWith('event:')) {
                        // Skip event line, data is on next iteration
                    } else if (line.startsWith('data:')) {
                        try {
                            const data = JSON.parse(line.slice(5).trim())
                            if (data.delta) {
                                callbacks.onMessage(data.delta)
                            }
                            if (data.message_id !== undefined && data.content !== undefined) {
                                callbacks.onDone(data.message_id, data.content)
                            }
                            if (data.msg) {
                                callbacks.onError(data.msg)
                            }
                        } catch (e) {
                            // Ignore parse errors for incomplete data
                        }
                    }
                }
            }
        })
        .catch((error) => {
            if (error.name !== 'AbortError') {
                callbacks.onError(error.message || 'Network error')
            }
        })

    // Return cancel function
    return () => controller.abort()
}
