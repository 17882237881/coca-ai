import axios from 'axios';

// Create Axios Instance
const request = axios.create({
    baseURL: 'http://localhost:8080', // Backend URL
    timeout: 5000,
});

// Request Interceptor: Inject Token
request.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('access_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response Interceptor: Handle Refresh Token
request.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;

        // Check if 401 and retry not already attempted
        if (error.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;

            const refreshToken = localStorage.getItem('refresh_token');
            if (refreshToken) {
                try {
                    // Call Refresh Endpoint with Silent Interceptor Bypass (create new instance or just raw axios if needed, but here simple post is fine if url is distinct)
                    // Actually better to use raw axios to avoid infinite loop if refresh fails 401 again
                    const { data } = await axios.post('http://localhost:8080/users/refresh_token', {
                        refresh_token: refreshToken
                    });

                    if (data.code === 200) {
                        // Update Local Storage
                        localStorage.setItem('access_token', data.data.access_token);
                        localStorage.setItem('refresh_token', data.data.refresh_token);

                        // Update Header and Retry
                        originalRequest.headers.Authorization = `Bearer ${data.data.access_token}`;
                        return request(originalRequest);
                    }
                } catch (refreshError) {
                    // Refresh Failed -> Logout
                    console.error("Refresh Token Failed", refreshError);
                    localStorage.removeItem('access_token');
                    localStorage.removeItem('refresh_token');
                    window.location.href = '/login';
                }
            } else {
                // No Refresh Token -> Logout
                window.location.href = '/login';
            }
        }
        return Promise.reject(error);
    }
);

export default request;
