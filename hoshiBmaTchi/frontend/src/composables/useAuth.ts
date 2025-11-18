import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { setAuthHeader } from "../services/apiService"; // Make sure to export this from apiService.ts


const isLoggedIn = ref(false);

export function useAuth(){
    const router = useRouter();

    const checkLoginState = () => {
      const token = localStorage.getItem("accessToken");
      isLoggedIn.value = !!token;
      if (token) {
        setAuthHeader(token); // Set header on app load
      }
    };

    const logout = () => {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
        setAuthHeader('');
        isLoggedIn.value = false;
        router.push({name: 'login'});
    };

    return {
        isLoggedIn,
        logout,
        checkLoginState,
    }
}