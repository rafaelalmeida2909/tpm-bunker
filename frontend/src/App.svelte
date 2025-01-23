<script>
  import { onDestroy, onMount } from "svelte";

  import Download from "svelte-icons/fa/FaDownload.svelte";
  import Lock from "svelte-icons/fa/FaLock.svelte";
  import ShieldCheck from "svelte-icons/fa/FaShieldAlt.svelte";
  import Sync from "svelte-icons/fa/FaSync.svelte";
  import Upload from "svelte-icons/fa/FaUpload.svelte";
  import Shield from "svelte-icons/fa/FaUserShield.svelte";
  import {
    AuthLogin,
    CheckConnection,
    CheckTPMPresence,
    GetOperations,
    InitializeDevice,
    IsDeviceInitialized,
  } from "../wailsjs/go/main/App";
  import FileEncryptionModal from "./components/FileEncryptionModal.svelte";

  // Estado do sistema
  let systemState = {
    tpmAvailable: false,
    deviceInitialized: false,
    apiConnected: true,
    authenticated: false,
    checking: true,
    initializationFailed: false,
  };

  let showEncryptionModal = false;
  let connectionCheckInterval;
  let initializationRetryInterval;

  let files = [];

  async function getOperations() {
  if (!systemState.authenticated) return;

  try {
    const response = await GetOperations();
    if (!response) {
      files = [];
      return;
    }

    // @ts-ignore
    const decodedStr = atob(response);
    const parsedData = JSON.parse(decodedStr);
    
    if (!Array.isArray(parsedData)) {
      console.error("Parsed data is not an array:", parsedData);
      files = [];
      return;
    }

    files = parsedData.filter(file => file && file.file_name);
  } catch (error) {
    console.error("Error in getOperations:", error);
    files = [];
  }
}

  function formatFileSize(size) {
    if (!size) return "0 B";
    const kb = size * 1024;
    if (kb < 1024) return `${kb.toFixed(2)} KB`;
    return `${size.toFixed(2)} MB`;
  }

  // Função para verificar conexão com a API
  async function checkAPIConnection() {
    try {
      systemState.apiConnected = await CheckConnection();
    } catch (error) {
      console.error("Erro ao verificar conexão:", error);
      systemState.apiConnected = false;
    }
  }

  // Função para inicializar o dispositivo
  async function initializeDeviceIfNeeded() {
    if (systemState.tpmAvailable && !systemState.deviceInitialized) {
      try {
        const deviceInfo = await InitializeDevice();
        if (deviceInfo) {
          const isAuthenticated = await AuthLogin();
          systemState = {
            ...systemState,
            deviceInitialized: true,
            initializationFailed: false,
            authenticated: isAuthenticated,
          };
          if (isAuthenticated) {
            await getOperations(); // Adicionado aqui
          }
          console.log("Device initialized successfully:", deviceInfo);
          if (initializationRetryInterval) {
            clearInterval(initializationRetryInterval);
            initializationRetryInterval = null;
          }
        }
      } catch (error) {
        console.error("Error initializing device:", error);
        systemState = {
          ...systemState,
          initializationFailed: true,
        };
        if (!initializationRetryInterval) {
          console.log("Setting up initialization retry interval");
          initializationRetryInterval = setInterval(
            initializeDeviceIfNeeded,
            10000,
          );
        }
      }
    }
  }

  // Verificação inicial do sistema
  async function checkSystemState() {
    try {
      const [tpmAvailable, deviceInitialized] = await Promise.all([
        CheckTPMPresence(),
        IsDeviceInitialized(),
      ]);

      systemState = {
        ...systemState,
        tpmAvailable,
        deviceInitialized,
        checking: false,
      };

      // Se TPM está disponível mas não inicializado, tenta inicializar
      if (tpmAvailable && !deviceInitialized) {
        await initializeDeviceIfNeeded();
      }

      // Primeira verificação de conexão
      await checkAPIConnection();
    } catch (error) {
      console.error("Error checking system state:", error);
      systemState.checking = false;
    }
  }

  onMount(() => {
    // Inicia verificação do sistema após 2 segundos
    setTimeout(() => {
      checkSystemState();
    }, 2000);

    // Inicia verificação periódica da conexão API (a cada 2 minutos)
    connectionCheckInterval = setInterval(checkAPIConnection, 2 * 60 * 1000);
  });

  onDestroy(() => {
    // Limpa os intervalos quando o componente é destruído
    if (connectionCheckInterval) {
      clearInterval(connectionCheckInterval);
    }
    if (initializationRetryInterval) {
      clearInterval(initializationRetryInterval);
    }
  });

  function encryptFile() {
    showEncryptionModal = true;
    getOperations();
  }

  function decryptFile(id) {
    console.log("Descriptografando arquivo...", id);
  }
</script>

<div class="app-container">
  {#if systemState.checking}
    <div class="p-6">
      <div class="alert alert-info">
        <h3 class="font-bold">Verificando estado do sistema...</h3>
        <p>
          Aguarde enquanto verificamos a disponibilidade do TPM e a conexão com
          a API.
        </p>
      </div>
    </div>
  {:else if !systemState.tpmAvailable || !systemState.apiConnected}
    <div class="p-6">
      <div class="alert alert-error">
        <h3 class="font-bold">Erro de Inicialização</h3>
        <p>
          {#if !systemState.tpmAvailable}
            TPM não está disponível. Verifique se seu dispositivo possui TPM e
            se está ativado no BIOS.
          {:else}
            Não foi possível conectar à API. Verifique sua conexão com a
            internet.
          {/if}
        </p>
      </div>
    </div>
  {:else}
    <div class="p-6 space-y-6">
      <!-- Status do Sistema -->
      <div class="system-status">
        <h2 class="text-xl font-bold mb-4">Status do Sistema</h2>
        <p class="text-gray-600 mb-4">
          Estado atual dos componentes do sistema
        </p>

        <div class="status-grid">
          <div class="status-item">
            <div
              class="icon"
              class:icon-green={systemState.tpmAvailable}
              class:icon-red={!systemState.tpmAvailable}
            >
              <Shield />
            </div>
            <span
              >TPM: {systemState.tpmAvailable
                ? "Disponível"
                : "Indisponível"}</span
            >
          </div>

          <div class="status-item">
            <div
              class="icon"
              class:icon-green={systemState.deviceInitialized}
              class:icon-red={!systemState.deviceInitialized &&
                systemState.initializationFailed}
              class:icon-blue={!systemState.deviceInitialized &&
                !systemState.initializationFailed}
            >
              {#if !systemState.deviceInitialized && !systemState.initializationFailed}
                <div class="icon-spin">
                  <Sync />
                </div>
              {:else}
                <ShieldCheck />
              {/if}
            </div>
            <span>
              Dispositivo:
              {#if !systemState.deviceInitialized && !systemState.initializationFailed}
                Inicializando...
              {:else if systemState.deviceInitialized}
                Inicializado
              {:else}
                Falha na Inicialização
              {/if}
            </span>
          </div>

          <div class="status-item">
            <div
              class="icon"
              class:icon-green={systemState.apiConnected}
              class:icon-red={!systemState.apiConnected}
            >
              <Sync />
            </div>
            <span
              >API: {systemState.apiConnected
                ? "Conectada"
                : "Desconectada"}</span
            >
          </div>

          <div class="status-item">
            <div
              class="icon"
              class:icon-green={systemState.authenticated}
              class:icon-red={!systemState.authenticated}
            >
              <Lock />
            </div>
            <span
              >Autenticação: {systemState.authenticated
                ? "Autenticado"
                : "Não Autenticado"}</span
            >
          </div>
        </div>
      </div>

      {#if systemState.authenticated}
        <!-- Gerenciador de Arquivos -->
        <div class="file-manager">
          <h2 class="text-xl font-bold mb-4">Arquivos Criptografados</h2>
          <p class="text-gray-600 mb-4">
            Gerencie seus arquivos protegidos pelo TPM
          </p>

          <div class="mb-6">
            <button class="btn btn-primary" on:click={encryptFile}>
              <div class="icon">
                <Upload />
              </div>
              Criptografar Arquivo
            </button>

            {#if showEncryptionModal}
              <FileEncryptionModal
                on:close={() => (showEncryptionModal = false)}
                on:fileEncrypted={() => getOperations()}
                isDeviceInitialized={systemState.deviceInitialized}
              />
            {/if}
          </div>

          <div class="border rounded-lg">
            <div class="file-header">
              <div>Nome</div>
              <div>Data</div>
              <div>Tamanho</div>
              <div>Ações</div>
            </div>

            {#each files as file (file.id)}
              <div class="file-row">
                <div>{file.file_name}</div>
                <div>{new Date(file.created_at).toLocaleString()}</div>
                <div>{formatFileSize(file.file_size)}</div>
                <div>
                  <button
                    class="btn btn-outline"
                    on:click={() => decryptFile(file.id)}
                  >
                    <div class="icon"><Download /></div>
                    Descriptografar
                  </button>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .app-container {
    @apply min-h-screen bg-gray-100;
  }

  .system-status {
    @apply bg-white rounded-lg shadow-md p-6 mb-6;
  }

  .status-grid {
    @apply grid grid-cols-2 gap-4;
  }

  .status-item {
    @apply flex items-center space-x-2;
  }

  .icon {
    @apply w-5 h-5;
  }

  .icon-green {
    @apply text-green-500;
  }

  .icon-red {
    @apply text-red-500;
  }

  .file-manager {
    @apply bg-white rounded-lg shadow-md p-6;
  }

  .file-header {
    @apply grid grid-cols-4 gap-4 p-4 bg-gray-50 border-b;
  }

  .file-row {
    @apply grid grid-cols-4 gap-4 p-4 border-b last:border-b-0;
  }

  .btn {
    @apply px-4 py-2 rounded-md flex items-center gap-2;
  }

  .btn-primary {
    @apply bg-blue-600 text-white hover:bg-blue-700;
  }

  .btn-outline {
    @apply border border-gray-300 hover:bg-gray-50;
  }

  .alert {
    @apply p-4 rounded-md mb-4;
  }

  .alert-info {
    @apply bg-blue-50 text-blue-700;
  }

  .alert-error {
    @apply bg-red-50 text-red-700;
  }

  .icon-blue {
    @apply text-blue-500;
  }

  .icon-spin {
    @apply animate-spin;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
