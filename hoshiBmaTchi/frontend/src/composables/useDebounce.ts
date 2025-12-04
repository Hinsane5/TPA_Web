import { ref, watch } from "vue";

export function useDebounce<T>(initialValue: T, delay = 300) {
  const value = ref(initialValue);
  const debouncedValue = ref(initialValue);
  let timeoutId: ReturnType<typeof setTimeout>;

  watch(value, (newValue) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => {
      debouncedValue.value = newValue as T;
    }, delay);
  });

  return { value, debouncedValue };
}
