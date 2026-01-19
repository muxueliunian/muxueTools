import axios from 'axios';

const client = axios.create({
    baseURL: '',
    timeout: 30000,
});

client.interceptors.response.use(
    (response) => {
        return response.data;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default client;
