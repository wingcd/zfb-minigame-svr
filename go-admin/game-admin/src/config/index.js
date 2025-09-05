export default {
  api: {
    baseURL: 'http://localhost:8080',  // Go admin-service 端口
    timeout: 10000
  },
  pagination: {
    pageSize: 20,
    pageSizes: [10, 20, 50, 100]
  }
}