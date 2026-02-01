<template>
  <div class="flex items-center justify-center min-h-screen bg-dark">
    <div class="w-full max-w-md p-8 bg-dark text-white text-center">
      <img src="https://upload.wikimedia.org/wikipedia/commons/0/04/ChatGPT_logo.svg" alt="Logo" class="w-12 h-12 mx-auto mb-5 invert" />
      <h1 class="text-3xl font-semibold mb-2">Create your account</h1>
      <div class="mb-6 text-sm text-gray-400">Sign up to get started.</div>
      
      <form @submit.prevent="handleSignup" class="space-y-4 text-left">
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
            minlength="8"
            class="w-full px-3 py-3 bg-surface border border-gray-600 rounded text-white focus:outline-none focus:border-primary transition-colors"
          />
        </div>
        
        <div>
          <label class="block text-xs font-bold text-gray-400 mb-1 uppercase">Confirm Password</label>
          <input 
            v-model="confirmPassword" 
            type="password" 
            required
            minlength="8"
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
        Already have an account? 
        <router-link to="/login" class="text-primary hover:underline">Log in</router-link>
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
const confirmPassword = ref('');
const loading = ref(false);

const handleSignup = async () => {
  if (password.value !== confirmPassword.value) {
    alert("Passwords do not match");
    return;
  }

  loading.value = true;
  try {
    const { data } = await request.post('/users/signup', {
      email: email.value,
      password: password.value,
      confirmPassword: confirmPassword.value
    });
    
    if (data.code === 200) {
      alert('Registration successful! Please login.');
      router.push('/login');
    } else {
      alert(data.msg || 'Registration failed');
    }
  } catch (err: any) {
    console.error(err);
    const msg = err.response?.data?.msg || 'An error occurred during registration';
    alert(msg);
  } finally {
    loading.value = false;
  }
};
</script>
