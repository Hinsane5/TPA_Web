import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import GoogleSignInPlugin from "vue3-google-signin";

const app = createApp(App);

app.use(GoogleSignInPlugin, {
  clientId: import.meta.env.VITE_GOOGLE_CLIENT_ID,
});

app.use(router);

app.mount("#app");
