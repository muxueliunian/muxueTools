import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

// Import fonts requires extra logic or just linking in index.html, 
// but assuming @fontsource packages are handled by vite
import '@fontsource/merriweather/400.css';
import '@fontsource/merriweather/700.css';
import '@fontsource/inter/400.css';
import '@fontsource/inter/500.css';
import '@fontsource/inter/600.css';

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
