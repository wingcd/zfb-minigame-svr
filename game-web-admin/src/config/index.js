export default {
  api: {
    baseURL: process.env.NODE_ENV === 'production' 
      ? 'https://your-alipay-function-domain.com' 
      : 'http://localhost:3000/api',
    timeout: 10000
  },
  pagination: {
    pageSize: 20,
    pageSizes: [10, 20, 50, 100]
  }
}