const currentEnv = import.meta.env.VITE_ENV || (import.meta.env.PROD ? 'PROD' : 'DEV')
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ''

const BASE_URL_MAP = {
  DEV: '',
  PROD: '',
  GP_PAGES: '',
  GITEE_PAGES: '/douyin',
  UNI: ''
}

export default {
  baseUrl: apiBaseUrl,
  imgPath: '/imgs/',
  filePreview: `${apiBaseUrl}/static/uploads/`
}

export const IS_SUB_DOMAIN = ['GITEE_PAGES', 'GP_PAGES'].includes(currentEnv)
export const IS_GITEE_PAGES = ['GITEE_PAGES'].includes(currentEnv)
export const BASE_URL = BASE_URL_MAP[currentEnv] || ''
export const IMG_URL = `${BASE_URL}/images/`
export const FILE_URL = `${BASE_URL}/data/`
export const IS_DEV = import.meta.env.DEV
