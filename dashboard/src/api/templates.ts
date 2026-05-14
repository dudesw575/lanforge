import api from "./client";

export interface TemplatePort {
  host: number;
  container: number;
}

export interface TemplateVolume {
  name: string;
  container: string;
}

export interface Template {
  id: string;
  name: string;
  description: string;
  image: string;
  env?: Record<string, string>;
  ports?: TemplatePort[];
  volumes?: TemplateVolume[];
  restart?: string;
}

export async function listTemplates(): Promise<Template[]> {
  const res = await api.get<Template[]>("/templates");
  return res.data;
}

export async function getTemplate(id: string): Promise<Template> {
  const res = await api.get<Template>(`/templates/${id}`);
  return res.data;
}

export async function createTemplate(data: Template): Promise<Template> {
  const res = await api.post<Template>("/templates", data, {
    headers: { "Content-Type": "application/json" },
  });
  return res.data;
}

export async function uploadTemplateYaml(file: File): Promise<Template> {
  const formData = new FormData();
  formData.append("file", file);
  const res = await api.post<Template>("/templates", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return res.data;
}

export async function updateTemplate(id: string, data: Partial<Template>): Promise<Template> {
  const res = await api.put<Template>(`/templates/${id}`, data, {
    headers: { "Content-Type": "application/json" },
  });
  return res.data;
}

export async function deleteTemplate(id: string): Promise<void> {
  await api.delete(`/templates/${id}`);
}

export async function deployTemplate(id: string, name: string): Promise<any> {
  const res = await api.post(
    `/templates/${id}/deploy?name=${encodeURIComponent(name)}`,
  );
  return res.data;
}