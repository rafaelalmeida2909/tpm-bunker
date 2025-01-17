<script>
  import { onMount } from 'svelte';
  import Download from 'svelte-icons/fa/FaDownload.svelte';
  import Lock from 'svelte-icons/fa/FaLock.svelte';

  import ShieldCheck from 'svelte-icons/fa/FaShieldAlt.svelte';
  import Sync from 'svelte-icons/fa/FaSync.svelte';
  import Upload from 'svelte-icons/fa/FaUpload.svelte';
  import Shield from 'svelte-icons/fa/FaUserShield.svelte';

  // Estado do sistema
  let systemState = {
    tpmAvailable: false,
    deviceInitialized: false,
    apiConnected: false,
    authenticated: false,
    checking: true
  };

  // Lista de arquivos
  let files = [
    { id: 1, name: "documento.pdf", date: "2024-01-16", size: "2.4 MB" },
    { id: 2, name: "contrato.docx", date: "2024-01-15", size: "1.1 MB" }
  ];

  // Simulação da verificação inicial
  onMount(() => {
    setTimeout(() => {
      systemState = {
        tpmAvailable: true,
        deviceInitialized: true,
        apiConnected: true,
        authenticated: true,
        checking: false
      };
    }, 2000);
  });

  // Funções de ação
  function encryptFile() {
    // Implementar lógica de criptografia
    console.log('Encriptando arquivo...');
  }

  function decryptFile(id) {
    // Implementar lógica de descriptografia
    console.log('Descriptografando arquivo...', id);
  }
</script>

<style lang="postcss">
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
</style>

{#if systemState.checking}
  <div class="p-6">
    <div class="alert alert-info">
      <h3 class="font-bold">Verificando estado do sistema...</h3>
      <p>Aguarde enquanto verificamos a disponibilidade do TPM e a conexão com a API.</p>
    </div>
  </div>
{:else if !systemState.tpmAvailable || !systemState.apiConnected}
  <div class="p-6">
    <div class="alert alert-error">
      <h3 class="font-bold">Erro de Inicialização</h3>
      <p>
        {#if !systemState.tpmAvailable}
          TPM não está disponível. Verifique se seu dispositivo possui TPM e se está ativado no BIOS.
        {:else}
          Não foi possível conectar à API. Verifique sua conexão com a internet.
        {/if}
      </p>
    </div>
  </div>
{:else}
  <div class="p-6 space-y-6">
    <!-- Status do Sistema -->
    <div class="system-status">
      <h2 class="text-xl font-bold mb-4">Status do Sistema</h2>
      <p class="text-gray-600 mb-4">Estado atual dos componentes do sistema</p>
      
      <div class="status-grid">
        <div class="status-item">
          <div class="icon" class:icon-green={systemState.tpmAvailable} class:icon-red={!systemState.tpmAvailable}>
            <Shield />
          </div>
          <span>TPM: {systemState.tpmAvailable ? 'Disponível' : 'Indisponível'}</span>
        </div>

        <div class="status-item">
          <div class="icon" class:icon-green={systemState.deviceInitialized} class:icon-red={!systemState.deviceInitialized}>
            <ShieldCheck />
          </div>
          <span>Dispositivo: {systemState.deviceInitialized ? 'Inicializado' : 'Não Inicializado'}</span>
        </div>

        <div class="status-item">
          <div class="icon" class:icon-green={systemState.apiConnected} class:icon-red={!systemState.apiConnected}>
            <Sync />
          </div>
          <span>API: {systemState.apiConnected ? 'Conectada' : 'Desconectada'}</span>
        </div>

        <div class="status-item">
          <div class="icon" class:icon-green={systemState.authenticated} class:icon-red={!systemState.authenticated}>
            <Lock />
          </div>
          <span>Autenticação: {systemState.authenticated ? 'Autenticado' : 'Não Autenticado'}</span>
        </div>
      </div>
    </div>

    {#if systemState.authenticated}
      <!-- Gerenciador de Arquivos -->
      <div class="file-manager">
        <h2 class="text-xl font-bold mb-4">Arquivos Criptografados</h2>
        <p class="text-gray-600 mb-4">Gerencie seus arquivos protegidos pelo TPM</p>

        <div class="mb-6">
          <button class="btn btn-primary" on:click={encryptFile}>
            <div class="icon">
              <Upload />
            </div>
            Criptografar Arquivo
          </button>
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
              <div>{file.name}</div>
              <div>{file.date}</div>
              <div>{file.size}</div>
              <div>
                <button class="btn btn-outline" on:click={() => decryptFile(file.id)}>
                  <div class="icon">
                    <Download />
                  </div>
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