/**
 * LogiFlows Client-Side Firebase Configuration
 * 
 * SECURITY NOTICE:
 * Only public client keys are referenced here.
 * NEVER add private keys, client_email, or service account secrets to this file!
 */

export interface FirebaseClientConfig {
  apiKey: string;
  authDomain: string;
  projectId: string;
  storageBucket: string;
  messagingSenderId: string;
  appId: string;
}

export const firebaseConfig: FirebaseClientConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY || 'AIzaSyMockKeyForDevOnly_Placeholder',
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN || 'logiflows-platform.firebaseapp.com',
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID || 'logiflows-platform',
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET || 'logiflows-platform.appspot.com',
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID || '123456789012',
  appId: import.meta.env.VITE_FIREBASE_APP_ID || '1:123456789012:web:abcdef123456',
};

/**
 * Check if live Firebase Client SDK is configured with non-placeholder keys
 */
export const isFirebaseConfigured = (): boolean => {
  return (
    Boolean(import.meta.env.VITE_FIREBASE_API_KEY) &&
    import.meta.env.VITE_FIREBASE_API_KEY !== 'AIzaSyMockKeyForDevOnly_Placeholder'
  );
};
