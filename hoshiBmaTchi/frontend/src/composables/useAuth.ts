import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { setAuthHeader } from "../services/apiService"; 
import { jwtDecode } from "jwt-decode";

const token = ref<string | null>(localStorage.getItem("accessToken"));
const user = ref<any>(null);
const isLoggedIn = ref(false);

interface JWTPayload {
  user_id: string;
  email: string;
  role: string;
  exp: number;
}

interface User {
  id: string;
  email: string;
}

export function useAuth(){
    const router = useRouter();

    const checkLoginState = () => {
      const storedToken = localStorage.getItem("accessToken");
      token.value = storedToken;
      isLoggedIn.value = !!storedToken;

      if (storedToken) {
        setAuthHeader(storedToken);
        try {
          const decoded = jwtDecode<JWTPayload>(storedToken);
          user.value = {
            id: decoded.user_id,
            email: decoded.email,
            role: decoded.role,
          };
        } catch (error) {
          console.error("Invalid token:", error);
          logout();
        }
      } else {
        user.value = null;
      }
    };

    const logout = () => {
      localStorage.removeItem("accessToken");
      localStorage.removeItem("refreshToken");
      setAuthHeader("");
      isLoggedIn.value = false;
      user.value = null;
      token.value = null;
      router.push({ name: "login" });
    };

    if (!user.value) {
      checkLoginState();
    }

    return {
      token, 
      user,
      isLoggedIn,
      logout,
      checkLoginState,
    };
}