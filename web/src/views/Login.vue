<template>
  <div class="flex items-center justify-center min-h-screen bg-dark">
    <div class="w-full max-w-md p-8 bg-dark text-white text-center">
      <img src="https://upload.wikimedia.org/wikipedia/commons/0/04/ChatGPT_logo.svg" alt="Logo" class="w-12 h-12 mx-auto mb-5 invert" />
      <h1 class="text-3xl font-semibold mb-2">Welcome back</h1>
      <div class="mb-6 text-sm text-gray-400">Please enter your details to sign in.</div>
      
      <form @submit.prevent="handleLogin" class="space-y-4 text-left">
        <div>
          <label class="block text-xs font-bold text-gray-400 mb-1 uppercase">Email address</label>
          <input 
            v-model="email" 
            type="email" 
            required
            class="w-full px-3 py-3 bg-surface border border-gray-600 rounded text-white focus:outline-none focus:border-primary transition-colors"
          />
        </div>
        
        <div>
          <label class="block text-xs font-bold text-gray-400 mb-1 uppercase">Password</label>
          <input 
            v-model="password" 
            type="password" 
            required
            class="w-full px-3 py-3 bg-surface border border-gray-600 rounded text-white focus:outline-none focus:border-primary transition-colors"
          />
        </div>
        
        <button 
          v-if="!loading"
          type="submit" 
          class="w-full py-3 bg-primary hover:bg-green-600 text-white font-medium rounded transition-colors"
        >
          Continue
        </button>
        <button 
          v-else
          disabled
          class="w-full py-3 bg-green-800 text-gray-300 font-medium rounded cursor-not-allowed"
        >
          Loading...
        </button>
      </form>
      
      <div class="mt-4 text-sm">
        Don't have an account? 
        <router-link to="/signup" class="text-primary hover:underline">Sign up</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import request from '../api/request';

const router = useRouter();
const email = ref('');
const password = ref('');
const loading = ref(false);

const handleLogin = async () => {
  loading.value = true;
  try {
    const { data } = await request.post('/users/login', {
      email: email.value,
      password: password.value
    });
    
    if (data.code === 200) {
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      router.push('/chat');
    } else {
      alert(data.msg || 'Login failed');
    }
  } catch (err) {
    console.error(err);
    alert('An error occurred during login');
  } finally {
    loading.value = false;
  }
};
</script>
