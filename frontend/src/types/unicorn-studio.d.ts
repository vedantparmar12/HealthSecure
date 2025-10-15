interface UnicornStudioInstance {
  isInitialized: boolean;
  init: () => void;
}

interface Window {
  UnicornStudio?: UnicornStudioInstance;
}
