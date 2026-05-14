<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRouter } from "vue-router";
import Modal from "../components/Modal.vue";
import ContainerLogs from "../components/ContainerLogs.vue";
import {
  listTemplates,
  createTemplate,
  uploadTemplateYaml,
  updateTemplate,
  deleteTemplate,
  deployTemplate as apiDeployTemplate,
  type Template,
} from "../api/templates";
import { templateToYaml, yamlToTemplate } from "../utils/yaml";

const router = useRouter();
const authStore = useAuthStore();
const templates = ref<Template[]>([]);

// Deploy state
const deployModalVisible = ref(false);
const deployTemplate = ref<Template | null>(null);
const deployName = ref("");
const deployResult = ref<{
  streamId: string;
  containerId: string | null;
  containerName: string | null;
  status: "deploying" | "success" | "failed";
  messages: string[];
} | null>(null);
const logViewerVisible = ref(false);
const logViewerContainerId = ref<string | null>(null);
let deployWS: WebSocket | null = null;

// Create / upload / edit state
const templateModalVisible = ref(false);
const templateFormMode = ref<"create" | "upload" | "edit">("create");
const editView = ref<"form" | "yaml">("form");
const editingTemplateId = ref<string | null>(null);
const editYaml = ref("");
const newTemplate = ref({
  id: "",
  name: "",
  description: "",
  image: "",
  env: {} as Record<string, string>,
  ports: [] as { host: number; container: number }[],
  volumes: [] as { name: string; container: string }[],
  restart: "",
});
const envKey = ref("");
const envVal = ref("");
const uploadFile = ref<File | null>(null);
const templateError = ref("");
const templateSuccess = ref("");
const deleteConfirmTemplate = ref<Template | null>(null);

const isAdmin = computed(() => {
  const roles = authStore.user?.resource_access?.["lan-control-plane"]?.roles ?? [];
  return roles.includes("admin");
});
const canDeploy = computed(() => {
  const roles = authStore.user?.resource_access?.["lan-control-plane"]?.roles ?? [];
  return roles.includes("admin") || roles.includes("operator");
});

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    return router.push("/");
  }
  await fetchTemplates();
});

onUnmounted(() => {
  deployWS?.close();
});

const fetchTemplates = async () => {
  try {
    templates.value = await listTemplates();
  } catch (err) {
    console.error(err);
  }
};

// Deploy
const openDeployModal = (tmpl: Template) => {
  deployTemplate.value = tmpl;
  deployName.value = `${tmpl.id}-${Math.floor(Math.random() * 1000)}`;
  deployModalVisible.value = true;
};

const submitDeploy = async () => {
  if (!deployTemplate.value || !deployName.value.trim()) return;
  const streamId = deployName.value;

  deployModalVisible.value = false;
  deployResult.value = {
    streamId,
    containerId: null,
    containerName: null,
    status: "deploying",
    messages: ["🔌 Connecting to deployment stream..."],
  };

  // Set up the WebSocket (with auth token) BEFORE awaiting connection so we never miss messages
  deployWS = new WebSocket(
    `ws://localhost:8080/deploy/stream?id=${streamId}&token=${authStore.token}`,
  );

  // Always set the onmessage handler immediately so no broadcasts are lost
  deployWS.onmessage = (e) => {
    if (!deployResult.value) return;
    deployResult.value.messages.push(e.data);

    if (e.data.includes("❌")) {
      deployResult.value.status = "failed";
    } else if (
      e.data.includes("complete") ||
      e.data.includes("🎉") ||
      e.data.includes("successfully")
    ) {
      deployResult.value.status = "success";
      const nameMatch = e.data.match(/Container ID:\s*(\S+)/i);
      if (nameMatch) {
        deployResult.value.containerId = nameMatch[1];
      }
    }
  };

  // Wait briefly for WebSocket to open before hitting the deploy API
  const wsConnected = await new Promise<boolean>((resolve) => {
    if (!deployWS) return resolve(false);
    const timer = setTimeout(() => resolve(false), 2000);
    deployWS.onopen = () => { clearTimeout(timer); resolve(true); };
    deployWS.onerror = () => { clearTimeout(timer); resolve(false); };
  });

  try {
    const result = await apiDeployTemplate(deployTemplate.value!.id, streamId);
    if (deployResult.value) {
      deployResult.value.containerId = result.containerId || null;
      deployResult.value.containerName = result.name || streamId;
      // Always show success from the API response, regardless of WebSocket state
      const shortId = result.containerId ? result.containerId.substring(0, 12) : "";
      const successMsg = wsConnected
        ? `✅ Deployment complete — container ${shortId} started`
        : `✅ Deployment complete — container ${shortId} started (live progress unavailable)`;
      // Only push if the WS messages didn't already mark it as success
      if (deployResult.value.status !== "success") {
        deployResult.value.messages.push(successMsg);
        deployResult.value.status = "success";
      }
    }
  } catch (err: any) {
    if (deployResult.value) {
      const errMsg = err.response?.data || err.message || "Unknown error";
      deployResult.value.messages.push(`❌ Deployment failed: ${errMsg}`);
      deployResult.value.status = "failed";
    }
  }
};

const dismissDeployResult = () => {
  deployResult.value = null;
  deployWS?.close();
  deployWS = null;
};

const openContainerLogs = (containerId: string, name: string) => {
  logViewerContainerId.value = containerId;
  logViewerVisible.value = true;
};

const closeContainerLogs = () => {
  logViewerVisible.value = false;
  logViewerContainerId.value = null;
};

// Create form
const openCreateForm = () => {
  templateFormMode.value = "create";
  newTemplate.value = {
    id: "",
    name: "",
    description: "",
    image: "",
    env: {},
    ports: [],
    volumes: [],
    restart: "",
  };
  envKey.value = "";
  envVal.value = "";
  templateError.value = "";
  templateSuccess.value = "";
  templateModalVisible.value = true;
};

const openUploadForm = () => {
  templateFormMode.value = "upload";
  uploadFile.value = null;
  templateError.value = "";
  templateSuccess.value = "";
  templateModalVisible.value = true;
};

const openEditForm = (tmpl: Template) => {
  templateFormMode.value = "edit";
  editingTemplateId.value = tmpl.id;
  editView.value = "form"; // Start with form view when editing
  // Populate form with template data
  newTemplate.value = {
    id: tmpl.id,
    name: tmpl.name,
    description: tmpl.description,
    image: tmpl.image,
    env: tmpl.env ?? {},
    ports: tmpl.ports ?? [],
    volumes: tmpl.volumes ?? [],
    restart: tmpl.restart ?? "",
  };
  // Also set the YAML view for switching
  editYaml.value = templateToYaml(newTemplate.value);
  templateError.value = "";
  templateSuccess.value = "";
  templateModalVisible.value = true;
};

const addEnv = () => {
  if (!envKey.value.trim()) return;
  newTemplate.value.env = {
    ...newTemplate.value.env,
    [envKey.value.trim()]: envVal.value,
  };
  envKey.value = "";
  envVal.value = "";
};

const removeEnv = (key: string) => {
  const next = { ...newTemplate.value.env };
  delete next[key];
  newTemplate.value.env = next;
};

const addPort = () => {
  newTemplate.value.ports.push({ host: 0, container: 0 });
};

const removePort = (i: number) => {
  newTemplate.value.ports.splice(i, 1);
};

const addVolume = () => {
  newTemplate.value.volumes.push({ name: "", container: "" });
};

const removeVolume = (i: number) => {
  newTemplate.value.volumes.splice(i, 1);
};

const submitCreateTemplate = async () => {
  templateError.value = "";
  templateSuccess.value = "";
  const t = newTemplate.value;
  if (!t.id || !t.name || !t.image) {
    templateError.value = "ID, Name, and Image are required.";
    return;
  }
  try {
    if (templateFormMode.value === "create") {
      await createTemplate({
        id: t.id,
        name: t.name,
        description: t.description,
        image: t.image,
        env: Object.keys(t.env).length > 0 ? t.env : undefined,
        ports: t.ports.length > 0 ? t.ports : undefined,
        volumes: t.volumes.length > 0 ? t.volumes : undefined,
        restart: t.restart || undefined,
      });
      templateSuccess.value = `Template "${t.name}" created!`;
    } else if (templateFormMode.value === "edit" && editingTemplateId.value) {
      await updateTemplate(editingTemplateId.value, {
        id: t.id,
        name: t.name,
        description: t.description,
        image: t.image,
        env: Object.keys(t.env).length > 0 ? t.env : undefined,
        ports: t.ports.length > 0 ? t.ports : undefined,
        volumes: t.volumes.length > 0 ? t.volumes : undefined,
        restart: t.restart || undefined,
      });
      templateSuccess.value = `Template "${t.name}" updated!`;
    }
    await fetchTemplates();
    setTimeout(() => { templateModalVisible.value = false; }, 1500);
  } catch (err: any) {
    templateError.value = err.response?.data || "Failed to save template.";
  }
};

const submitUploadTemplate = async () => {
  templateError.value = "";
  templateSuccess.value = "";
  if (!uploadFile.value) {
    templateError.value = "Please select a YAML file.";
    return;
  }
  try {
    const result = await uploadTemplateYaml(uploadFile.value);
    templateSuccess.value = `Template "${result.name}" uploaded!`;
    await fetchTemplates();
    setTimeout(() => { templateModalVisible.value = false; }, 1500);
  } catch (err: any) {
    templateError.value = err.response?.data || "Failed to upload template.";
  }
};

// Edit YAML view handling
const switchToYamlView = () => {
  if (templateFormMode.value === "edit" && editingTemplateId.value) {
    // Update the YAML from the current form state
    editYaml.value = templateToYaml(newTemplate.value);
    editView.value = "yaml";
  }
};

const switchToFormView = () => {
  if (templateFormMode.value === "edit" && editingTemplateId.value) {
    // Update the form from the YAML
    try {
      const templateObj = yamlToTemplate(editYaml.value);
      newTemplate.value = {
        id: templateObj.id,
        name: templateObj.name,
        description: templateObj.description,
        image: templateObj.image,
        env: templateObj.env ?? {},
        ports: templateObj.ports ?? [],
        volumes: templateObj.volumes ?? [],
        restart: templateObj.restart ?? "",
      };
      editView.value = "form";
    } catch (e) {
      templateError.value = "Invalid YAML: " + (e as Error).message;
    }
  }
};

const submitEditYaml = async () => {
  templateError.value = "";
  templateSuccess.value = "";
  if (!editingTemplateId.value) return;
  try {
    const templateObj = yamlToTemplate(editYaml.value);
    await updateTemplate(editingTemplateId.value, {
      id: templateObj.id,
      name: templateObj.name,
      description: templateObj.description,
      image: templateObj.image,
      env: templateObj.env ?? {},
      ports: templateObj.ports ?? [],
      volumes: templateObj.volumes ?? [],
      restart: templateObj.restart || undefined,
    });
    templateSuccess.value = `Template "${templateObj.name}" updated!`;
    await fetchTemplates();
    // Switch back to form view after successful update
    editView.value = "form";
    setTimeout(() => { templateModalVisible.value = false; }, 1500);
  } catch (err: any) {
    templateError.value = err.response?.data || "Failed to update template from YAML.";
  }
};

// Delete
const confirmDeleteTemplate = (tmpl: Template) => {
  deleteConfirmTemplate.value = tmpl;
};

const executeDeleteTemplate = async () => {
  if (!deleteConfirmTemplate.value) return;
  try {
    await deleteTemplate(deleteConfirmTemplate.value.id);
    await fetchTemplates();
  } catch (err) {
    console.error("Failed to delete template:", err);
  }
  deleteConfirmTemplate.value = null;
};

const cancelDeleteTemplate = () => {
  deleteConfirmTemplate.value = null;
};

const envEntries = computed(() => Object.entries(newTemplate.value.env));
const fileInput = ref<HTMLInputElement | null>(null);
</script>

<template>
  <div class="bg-bg text-text transition-colors duration-300">
    <div class="container mx-auto py-8 px-4">
      
      <div class="flex items-center justify-between mb-8">
        <div>
          <h1 class="text-3xl font-bold flex items-center gap-3">
            <span class="text-primary">📜</span> Templates
          </h1>
          <p class="text-sm opacity-50 mt-1">Manage game server deployment templates</p>
        </div>

        <div v-if="isAdmin" class="flex gap-2">
          <button
            @click="openCreateForm"
            class="bg-green-600 hover:bg-green-700 text-white px-5 py-2.5 rounded-lg font-bold shadow-md active:scale-95 transition-all"
          >
            + New Form
          </button>
          <button
            @click="openUploadForm"
            class="bg-blue-600 hover:bg-blue-700 text-white px-5 py-2.5 rounded-lg font-bold shadow-md active:scale-95 transition-all"
          >
            ↑ Upload YAML
          </button>
        </div>
      </div>

      <!-- Deploy Result Panel (replaces progress bar) -->
      <div
        v-if="deployResult"
        class="mb-6 border rounded-xl p-4"
        :class="{
          'bg-white/5 border-current/10': deployResult.status === 'deploying',
          'bg-green-900/20 border-green-700/40': deployResult.status === 'success',
          'bg-red-900/20 border-red-700/40': deployResult.status === 'failed',
        }"
      >
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-3">
            <div v-if="deployResult.status === 'deploying'" class="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full"></div>
            <span v-else class="text-lg">{{ deployResult.status === 'success' ? '✅' : '❌' }}</span>
            <span class="font-bold">{{ deployResult.containerName || deployResult.streamId }}</span>
          </div>
          <button @click="dismissDeployResult" class="text-xs opacity-40 hover:opacity-100 px-2 py-1 rounded hover:bg-white/10">&times; Dismiss</button>
        </div>

        <!-- Deployment logs (always visible during and after) -->
        <div class="space-y-0.5 text-sm font-mono max-h-48 overflow-y-auto bg-black/30 rounded-lg p-3">
          <div v-for="(msg, i) in deployResult.messages" :key="i" class="opacity-90 leading-relaxed">{{ msg }}</div>
        </div>

        <!-- Post-deploy actions (success only) -->
        <div v-if="deployResult.status === 'success' && deployResult.containerId" class="flex gap-3 mt-4">
          <router-link
            :to="'/dashboard'"
            class="bg-primary hover:bg-primary-hover text-navbar-text px-4 py-2 rounded-lg font-bold text-sm shadow-md transition-all"
          >
            Go to Dashboard
          </router-link>
          <button
            @click="openContainerLogs(deployResult.containerId!, deployResult.containerName || deployResult.streamId)"
            class="bg-white/10 hover:bg-white/20 text-current px-4 py-2 rounded-lg font-bold text-sm transition-all"
          >
            View Container Logs
          </button>
        </div>
      </div>

      <!-- Template Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        <div
          v-for="t in templates"
          :key="t.id"
          class="border border-current/10 p-5 rounded-xl bg-white/5 backdrop-blur-sm flex flex-col gap-3 hover:border-primary/30 transition-all shadow-sm"
        >
          <div class="flex justify-between items-start">
            <div>
              <div class="font-bold text-primary text-lg leading-tight">{{ t.name }}</div>
              <div class="text-sm opacity-60 mt-1">{{ t.description }}</div>
            </div>
            <span class="text-xs font-mono opacity-40 bg-white/5 px-2 py-0.5 rounded whitespace-nowrap">{{ t.id }}</span>
          </div>

          <div class="flex items-center gap-2 text-xs opacity-60 font-mono">
            <span class="bg-white/5 px-2 py-0.5 rounded">{{ t.image }}</span>
            <span v-if="t.ports?.length" class="opacity-40">|</span>
            <span v-if="t.ports?.length">{{ t.ports.length }} port(s)</span>
            <span v-if="t.volumes?.length" class="opacity-40">|</span>
            <span v-if="t.volumes?.length">{{ t.volumes.length }} volume(s)</span>
          </div>

          <div class="flex gap-2 mt-1">
            <button
              v-if="canDeploy"
              @click="openDeployModal(t)"
              class="flex-1 bg-primary hover:bg-primary-hover text-navbar-text px-4 py-2 rounded-lg font-bold text-sm shadow-md active:scale-95 transition-all"
            >
              Deploy
            </button>
            <button
              v-if="isAdmin"
              @click="confirmDeleteTemplate(t)"
              class="px-3 py-2 bg-red-600/80 hover:bg-red-600 text-white rounded-lg text-sm font-bold transition-all active:scale-95"
              title="Delete template"
            >
              ✕
            </button>
            <button
              v-if="isAdmin"
              @click="openEditForm(t)"
              class="px-3 py-2 bg-blue-600/80 hover:bg-blue-600 text-white rounded-lg text-sm font-bold transition-all active:scale-95"
              title="Edit template"
            >
              ✎
            </button>
          </div>
        </div>

        <div v-if="templates.length === 0" class="col-span-full text-center opacity-40 py-16">
          <div class="text-4xl mb-4">📜</div>
          <p class="text-lg">No templates yet.</p>
          <p v-if="isAdmin" class="text-sm mt-1">Click "New Form" or "Upload YAML" to add one.</p>
          <p v-else class="text-sm mt-1">Ask an admin to create templates for deployment.</p>
        </div>
      </div>
    </div>

    <!-- Deploy Modal -->
    <Teleport to="body">
      <Modal :show="deployModalVisible" @close="deployModalVisible = false">
        <template #header>
          <span class="text-navbar-text font-bold text-lg">Deploy: {{ deployTemplate?.name }}</span>
        </template>
        <template #body>
          <div class="py-6">
            <label class="block text-sm opacity-60 mb-2 font-medium">Instance Name (Unique ID)</label>
            <input
              v-model="deployName"
              class="w-full bg-white/5 border border-current/20 p-3 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none transition-all"
              placeholder="e.g. my-minecraft-server"
            />
          </div>
        </template>
        <template #extra-actions>
          <button
            @click="submitDeploy"
            class="w-full sm:w-auto bg-primary hover:bg-primary-hover text-navbar-text px-8 py-2 rounded-lg font-bold shadow-lg"
          >
            Confirm Launch
          </button>
        </template>
      </Modal>
    </Teleport>

    <!-- Container Logs Modal -->
    <Teleport to="body">
      <Modal :show="logViewerVisible" @close="closeContainerLogs">
        <template #header>
          <span class="text-navbar-text font-bold text-lg">Container Logs: {{ deployResult?.containerName }}</span>
        </template>
        <template #body>
          <div class="bg-black/40 rounded-lg border border-current/10 overflow-hidden" style="height: 60vh;">
            <ContainerLogs
              v-if="logViewerVisible && logViewerContainerId"
              :containerID="logViewerContainerId"
            />
          </div>
        </template>
      </Modal>
    </Teleport>

    <!-- Create / Upload / Edit Template Form Modal -->
    <Teleport to="body">
      <Modal :show="templateModalVisible" @close="templateModalVisible = false">
        <template #header>
          <span class="text-navbar-text font-bold text-lg">
            {{ templateFormMode === 'create' ? 'Create Template' : templateFormMode === 'upload' ? 'Upload Template YAML' : 'Edit Template' }}
          </span>
        </template>
        <template #body>
          <div class="overflow-y-auto">
            <div v-if="templateError" class="bg-red-900/30 border border-red-700/50 rounded-lg p-3 text-red-200 text-sm">{{ templateError }}</div>
            <div v-if="templateSuccess" class="bg-green-900/30 border border-green-700/50 rounded-lg p-3 text-green-200 text-sm">{{ templateSuccess }}</div>

            <!-- Form View -->
            <div v-if="templateFormMode !== 'upload' && editView === 'form'">
              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">ID *</label>
                <input v-model="newTemplate.id" class="w-full bg-white/5 border border-current/20 p-2.5 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="e.g. minecraft-paper" />
              </div>
              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Name *</label>
                <input v-model="newTemplate.name" class="w-full bg-white/5 border border-current/20 p-2.5 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="Minecraft Paper Server" />
              </div>
              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Description</label>
                <input v-model="newTemplate.description" class="w-full bg-white/5 border border-current/20 p-2.5 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="A high-performance Minecraft server" />
              </div>
              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Image *</label>
                <input v-model="newTemplate.image" class="w-full bg-white/5 border border-current/20 p-2.5 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="itzg/minecraft-server" />
              </div>
              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Restart Policy</label>
                <select v-model="newTemplate.restart" class="w-full bg-white/5 border border-current/20 p-2.5 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm">
                  <option value="">None</option>
                  <option value="always">always</option>
                  <option value="unless-stopped">unless-stopped</option>
                  <option value="on-failure">on-failure</option>
                </select>
              </div>

              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Environment Variables</label>
                <div class="flex gap-2 mb-2">
                  <input v-model="envKey" @keyup.enter="addEnv" class="flex-1 bg-white/5 border border-current/20 p-2 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="KEY" />
                  <input v-model="envVal" @keyup.enter="addEnv" class="flex-1 bg-white/5 border border-current/20 p-2 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="VALUE" />
                  <button @click="addEnv" class="bg-primary hover:bg-primary-hover text-navbar-text px-3 rounded-lg text-sm font-bold">+</button>
                </div>
                <div class="space-y-1">
                  <div v-for="[k, v] in envEntries" :key="k" class="flex items-center gap-2 bg-white/5 px-3 py-1.5 rounded-lg text-sm">
                    <span class="font-mono font-bold text-primary">{{ k }}</span>
                    <span class="opacity-60">=</span>
                    <span class="font-mono text-xs opacity-80">{{ v }}</span>
                    <button @click="removeEnv(k)" class="ml-auto text-red-400 hover:text-red-300 text-xs font-bold">✕</button>
                  </div>
                </div>
              </div>

              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Port Mappings</label>
                <div v-for="(p, i) in newTemplate.ports" :key="i" class="flex gap-2 mb-1">
                  <input v-model.number="p.host" type="number" class="w-24 bg-white/5 border border-current/20 p-2 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="Host" />
                  <span class="self-center opacity-40">:</span>
                  <input v-model.number="p.container" type="number" class="w-24 bg-white/5 border border-current/20 p-2 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="Container" />
                  <button @click="removePort(i)" class="text-red-400 hover:text-red-300 text-sm px-2">✕</button>
                </div>
                <button @click="addPort" class="text-primary text-sm font-bold hover:underline">+ Add Port</button>
              </div>

              <div>
                <label class="block text-sm opacity-60 mb-1 font-medium">Volumes</label>
                <div v-for="(v, i) in newTemplate.volumes" :key="i" class="flex gap-2 mb-1">
                  <input v-model="v.name" class="flex-1 bg-white/5 border border-current/20 p-2 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="Volume name" />
                  <input v-model="v.container" class="flex-1 bg-white/5 border border-current/20 p-2 rounded-lg text-current focus:ring-2 focus:ring-primary/50 outline-none text-sm" placeholder="Container path" />
                  <button @click="removeVolume(i)" class="text-red-400 hover:text-red-300 text-sm px-2">✕</button>
                </div>
                <button @click="addVolume" class="text-primary text-sm font-bold hover:underline">+ Add Volume</button>
              </div>
            </div>

            <!-- YAML View -->
            <div v-if="templateFormMode !== 'upload' && editView === 'yaml'">
              <div class="mb-4">
                <button @click="switchToFormView" class="text-sm text-primary hover:underline">
                  ← Switch to Form View
                </button>
              </div>
              <div class="border border-current/20 rounded-lg p-4">
                <textarea
                  v-model="editYaml"
                  class="w-full h-96 font-mono text-sm bg-white/5 border border-current/20 rounded-lg p-2 text-current focus:ring-2 focus-ring-primary/50 outline-none resize-none"
                  placeholder="Edit template YAML here..."
                ></textarea>
              </div>
              <div class="mt-2">
                <button
                  @click="submitEditYaml"
                  class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-bold"
                >
                  Save Changes
                </button>
              </div>
            </div>

            <!-- Upload View -->
            <div v-if="templateFormMode === 'upload'">
              <div class="border-2 border-dashed border-current/20 rounded-xl p-8 text-center hover:border-primary/50 transition-all cursor-pointer" @click="fileInput?.click()">
                <div v-if="!uploadFile" class="space-y-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-8 w-8 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                  </svg>
                  <p class="text-sm opacity-60">Click to select a YAML template file</p>
                </div>
                <div v-else class="space-y-2">
                  <p class="text-primary font-bold">{{ uploadFile.name }}</p>
                  <p class="text-xs opacity-40">{{ (uploadFile.size / 1024).toFixed(1) }} KB</p>
                  <button @click.stop="uploadFile = null" class="text-red-400 text-sm hover:underline">Remove</button>
                </div>
                <input ref="fileInput" type="file" accept=".yaml,.yml" class="hidden" @change="uploadFile = ($event.target as HTMLInputElement).files?.[0] || null" />
              </div>
            </div>
          </div>
        </template>
        <template #extra-actions>
          <div v-if="templateFormMode === 'create'">
            <button
              @click="submitCreateTemplate"
              class="w-full sm:w-auto bg-green-600 hover:bg-green-700 text-white px-8 py-2 rounded-lg font-bold shadow-lg"
            >
              Create Template
            </button>
          </div>
          <div v-if="templateFormMode === 'upload'">
            <button
              @click="submitUploadTemplate"
              :disabled="!uploadFile"
              class="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 disabled:opacity-30 disabled:cursor-not-allowed text-white px-8 py-2 rounded-lg font-bold shadow-lg"
            >
              Upload & Save
            </button>
          </div>
          <div v-if="templateFormMode === 'edit'">
            <div v-if="editView === 'form'">
              <div class="flex justify-between">
                <button
                  @click="switchToYamlView"
                  class="text-sm text-primary hover:underline"
                >
                  Switch to YAML View
                </button>
                <button
                  @click="submitCreateTemplate"
                  class="w-full sm:w-auto bg-green-600 hover:bg-green-700 text-white px-8 py-2 rounded-lg font-bold shadow-lg"
                >
                  Save Changes
                </button>
              </div>
            </div>
            <div v-if="editView === 'yaml'">
              <!-- The save button is already in the YAML view above -->
            </div>
          </div>
        </template>
      </Modal>
    </Teleport>

    <!-- Delete Template Confirmation Modal -->
    <Teleport to="body">
      <Modal :show="!!deleteConfirmTemplate" @close="cancelDeleteTemplate">
        <template #header><span class="text-white font-bold">Delete Template</span></template>
        <template #body>
          <div class="p-2 text-center">
            <p class="mb-2 text-lg">Delete template <strong>{{ deleteConfirmTemplate?.name }}</strong>?</p>
            <p class="text-sm opacity-60">This removes the template YAML file. Running deployments are not affected.</p>
          </div>
        </template>
        <template #extra-actions>
          <div class="flex justify-end gap-3 w-full">
            <button @click="cancelDeleteTemplate" class="px-6 py-2 bg-white/10 hover:bg-white/20 rounded font-bold">Cancel</button>
            <button @click="executeDeleteTemplate" class="px-6 py-2 bg-red-600 hover:bg-red-500 rounded font-bold">Delete</button>
          </div>
        </template>
      </Modal>
    </Teleport>
  </div>
</template>