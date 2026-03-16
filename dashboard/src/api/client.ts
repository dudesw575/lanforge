import axios, { type AxiosRequestConfig, type InternalAxiosRequestConfig } from "axios";
import { getToken } from "../auth/oidc";

const api = axios.create({ baseURL: "/api" });

api.interceptors.request.use(
  async (config: AxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
    if (!config.headers) {
      config.headers = {} as InternalAxiosRequestConfig["headers"];
    }

    try {
      const token = await getToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      } else {
        console.warn("No token available for request");
      }
    } catch (err) {
      console.warn("No token available for request", err);
    }

    return config as InternalAxiosRequestConfig;
  },
  (error) => Promise.reject(error) 
);

export default api;