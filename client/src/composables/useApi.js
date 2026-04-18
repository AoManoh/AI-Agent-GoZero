import { inject } from "vue";
import apiService, { API_INJECTION_KEY } from "../api/index.js";

export const useApi = () => inject(API_INJECTION_KEY, apiService);
