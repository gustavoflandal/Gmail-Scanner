/**
 * Utilitários para gerenciamento de armazenamento local
 * Ajuda a evitar erros de quota de localStorage
 */

const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  USER_EMAIL: 'user_email',
  USER_NAME: 'user_name',
};

const MAX_STORAGE_SIZE = 1024 * 1024; // 1MB como limite seguro

/**
 * Calcula o tamanho aproximado do localStorage em bytes
 */
export function getStorageSize() {
  let total = 0;
  for (let key in localStorage) {
    if (localStorage.hasOwnProperty(key)) {
      total += key.length + localStorage[key].length;
    }
  }
  return total;
}

/**
 * Obtém a porcentagem de uso do localStorage
 */
export function getStorageUsagePercent() {
  return (getStorageSize() / MAX_STORAGE_SIZE) * 100;
}

/**
 * Salva um item no localStorage de forma segura
 */
export function saveToStorage(key, value) {
  try {
    // Verificar tamanho antes de salvar
    const valueToStore = (key === STORAGE_KEYS.AUTH_TOKEN) ? value : JSON.stringify(value);
    const newSize = getStorageSize() + key.length + valueToStore.length;
    
    if (newSize > MAX_STORAGE_SIZE * 0.9) {
      console.warn(`Storage quota approaching. Current size: ${newSize} bytes`);
      // Tentar limpar dados antigos
      cleanupStorage();
    }

    localStorage.setItem(key, valueToStore);
  } catch (error) {
    if (error.name === 'QuotaExceededError' || error.code === 22) {
      console.error('Storage quota exceeded. Cleaning up old data...');
      cleanupStorage();
      try {
        const valueToStore = (key === STORAGE_KEYS.AUTH_TOKEN) ? value : JSON.stringify(value);
        localStorage.setItem(key, valueToStore);
      } catch (retryError) {
        console.error('Failed to save to storage after cleanup:', retryError);
        throw retryError;
      }
    } else {
      throw error;
    }
  }
}

/**
 * Obtém um item do localStorage
 */
export function getFromStorage(key) {
  try {
    const item = localStorage.getItem(key);
    if (!item) return null;
    
    // Se for o token de autenticação, retornar diretamente (JWT é string)
    if (key === STORAGE_KEYS.AUTH_TOKEN) {
      return item;
    }
    
    // Para outros valores, fazer parse JSON
    try {
      return JSON.parse(item);
    } catch {
      // Se falhar o parse, retornar como string
      return item;
    }
  } catch (error) {
    console.error(`Failed to retrieve ${key} from storage:`, error);
    return null;
  }
}

/**
 * Remove um item do localStorage
 */
export function removeFromStorage(key) {
  try {
    localStorage.removeItem(key);
  } catch (error) {
    console.error(`Failed to remove ${key} from storage:`, error);
  }
}

/**
 * Limpa o localStorage completamente
 */
export function clearStorage() {
  try {
    localStorage.clear();
    console.log('Storage cleared successfully');
  } catch (error) {
    console.error('Failed to clear storage:', error);
  }
}

/**
 * Remove dados antigos e desnecessários do localStorage
 */
export function cleanupStorage() {
  const keysToCheck = [
    'vite:moduleResolutionCache', // Vite cache
    'viteDeps', // Vite dependencies
    '__REDUX_DEVTOOLS_EXTENSION_COMPOSE__', // Redux devtools
  ];

  let cleaned = false;

  // Remover chaves conhecidas de cache
  keysToCheck.forEach(key => {
    if (localStorage.getItem(key)) {
      removeFromStorage(key);
      cleaned = true;
    }
  });

  // Se ainda ocupar muito espaço, remover dados menos importantes
  if (getStorageSize() > MAX_STORAGE_SIZE * 0.8) {
    // Manter apenas chaves de autenticação
    const importantKeys = Object.values(STORAGE_KEYS);
    const allKeys = Object.keys(localStorage);

    allKeys.forEach(key => {
      if (!importantKeys.includes(key)) {
        removeFromStorage(key);
        cleaned = true;
      }
    });
  }

  if (cleaned) {
    console.log(`Cleaned storage. New size: ${getStorageSize()} bytes`);
  }

  return cleaned;
}

/**
 * Salva token de autenticação
 */
export function saveAuthToken(token) {
  saveToStorage(STORAGE_KEYS.AUTH_TOKEN, token);
}

/**
 * Obtém token de autenticação
 */
export function getAuthToken() {
  return getFromStorage(STORAGE_KEYS.AUTH_TOKEN);
}

/**
 * Remove token de autenticação
 */
export function removeAuthToken() {
  removeFromStorage(STORAGE_KEYS.AUTH_TOKEN);
}

/**
 * Salva informações do usuário
 */
export function saveUserInfo(email, name) {
  saveToStorage(STORAGE_KEYS.USER_EMAIL, email);
  saveToStorage(STORAGE_KEYS.USER_NAME, name);
}

/**
 * Obtém informações do usuário
 */
export function getUserInfo() {
  return {
    email: getFromStorage(STORAGE_KEYS.USER_EMAIL),
    name: getFromStorage(STORAGE_KEYS.USER_NAME),
  };
}

/**
 * Remove informações do usuário
 */
export function removeUserInfo() {
  removeFromStorage(STORAGE_KEYS.USER_EMAIL);
  removeFromStorage(STORAGE_KEYS.USER_NAME);
}

/**
 * Monitora uso de storage (log periodicamente)
 */
export function startStorageMonitoring(intervalMs = 60000) {
  setInterval(() => {
    const usage = getStorageUsagePercent();
    if (usage > 50) {
      console.warn(`Storage usage: ${usage.toFixed(2)}%`);
    }
  }, intervalMs);
}
