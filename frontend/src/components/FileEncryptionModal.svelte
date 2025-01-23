<script>
  import { createEventDispatcher } from "svelte";
  import Sync from "svelte-icons/fa/FaSync.svelte";
  import { fade } from "svelte/transition";
  import {
    EncryptFile,
    IsDeviceInitialized,
    SelectFile,
  } from "../../wailsjs/go/main/App";
  export let isDeviceInitialized = false;
  const dispatch = createEventDispatcher();

  async function handleUpload() {
    if (!selectedFile) return;

    try {
      const isInitialized = await IsDeviceInitialized();
      if (!isInitialized) {
        dispatch("showToast", {
          message: "Aguarde a inicialização do dispositivo ser concluída.",
          type: "error",
        });
        return;
      }

      isUploading = true;
      uploadProgress = 0;

      const progressInterval = setInterval(() => {
        if (uploadProgress < 90) {
          uploadProgress += 10;
        }
      }, 500);

      // Chama a função EncryptFile do backend
      const result = await EncryptFile(selectedFile.path);

      uploadProgress = 100;
      clearInterval(progressInterval);

      dispatch("showToast", {
        message: "Arquivo criptografado com sucesso!",
        type: "success",
      });
      dispatch("handleStartLockAnimation"); // Add this line
      dispatch("fileEncrypted");
      dispatch("close");
    } catch (error) {
      console.error("Erro ao criptografar arquivo:", error);
      dispatch("showToast", {
        message: "Erro ao criptografar arquivo. Tente novamente.",
        type: "error",
      });
    } finally {
      isUploading = false;
      uploadProgress = 0;
    }
  }

  let selectedFile = null;
  let isUploading = false;
  let uploadProgress = 0;
  let showToast = false;
  let toastMessage = "";
  let toastType = "success";


  async function handleFileSelect() {
    try {
      const filePath = await SelectFile();
      console.log(filePath);
      if (filePath) {
        selectedFile = {
          // @ts-ignore
          name: filePath.split("\\").pop().split("/").pop(), // Pega só o nome do arquivo
          path: filePath,
        };
      }
    } catch (error) {
      console.error("Erro ao selecionar arquivo:", error);
      dispatch("showToast", {
        message: "Erro ao selecionar arquivo. Tente novamente.",
        type: "error",
      });
    }
  }
</script>

<div
  class="modal-backdrop fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center"
>
  <div class="modal-content bg-white rounded-lg p-6 w-96 space-y-4">
    <h3 class="text-xl font-bold">Selecionar Arquivo</h3>

    {#if selectedFile}
      <p class="text-sm text-gray-600">
        Arquivo selecionado: {selectedFile.name}
      </p>
    {:else}
      <p class="text-sm text-gray-600">Nenhum arquivo selecionado</p>
    {/if}

    <div class="flex justify-center">
      <button
        class="btn btn-outline cursor-pointer"
        on:click={handleFileSelect}
      >
        Escolher Arquivo
      </button>
    </div>

    <div class="flex justify-end space-x-2 mt-4">
      <button class="btn btn-outline" on:click={() => dispatch("close")}>
        Cancelar
      </button>
      <button
        class="btn btn-primary"
        disabled={!selectedFile || isUploading || !isDeviceInitialized}
        on:click={handleUpload}
      >
        {#if isUploading}
          <div class="loading-icon">
            <Sync />
          </div>
          Enviando...
        {:else}
          Enviar
        {/if}
      </button>
    </div>

    {#if isUploading}
      <div class="w-full bg-gray-200 rounded-full h-2.5">
        <div
          class="bg-blue-600 h-2.5 rounded-full"
          style="width: {uploadProgress}%"
        ></div>
      </div>
    {/if}
  </div>
</div>

{#if showToast}
  <div
    class="fixed top-4 right-4 p-4 rounded-lg shadow-lg text-white {toastType ===
    'success'
      ? 'bg-green-500'
      : 'bg-red-500'}"
    transition:fade={{ duration: 200 }}
  >
    <p>{toastMessage}</p>
  </div>
{/if}

<style lang="postcss">
  .modal-backdrop {
    z-index: 1000;
  }

  .modal-content {
    z-index: 1001;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .loading-icon {
    @apply w-4 h-4;
    animation: spin 1s linear infinite;
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

  .icon {
    @apply w-5 h-5;
  }
</style>
