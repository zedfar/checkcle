import React, { useEffect } from "react";
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter 
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { AlertConfiguration, alertConfigService } from "@/services/alertConfigService";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Switch } from "@/components/ui/switch";
import { Loader2, Copy } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "@/hooks/use-toast";
import { useLanguage } from "@/contexts/LanguageContext";

interface NotificationChannelDialogProps {
  open: boolean;
  onClose: (refreshList: boolean) => void;
  editingConfig: AlertConfiguration | null;
}


const baseSchema = z.object({
  notify_name: z.string().min(1, "Name is required"),
  notification_type: z.enum(["telegram", "webhook"]),
  enabled: z.boolean().default(true),
  service_id: z.string().default("global"),
  template_id: z.string().optional(),
});

const telegramSchema = baseSchema.extend({
  notification_type: z.literal("telegram"),
  telegram_chat_id: z.string().min(1, "Chat ID is required"),
  bot_token: z.string().min(1, "Bot token is required"),
  telegram_thread_id: z.string().optional(),
});

const webhookSchema = baseSchema.extend({
  notification_type: z.literal("webhook"),
  webhook_url: z.string().url("Must be a valid URL"),
  webhook_payload_template: z.string().optional(),
});

const formSchema = z.discriminatedUnion("notification_type", [
  telegramSchema,
  webhookSchema,
]);

type FormValues = z.infer<typeof formSchema>;

const notificationTypeOptions = [
  {
    value: "telegram",
    label: "Telegram",
    description: "Send notifications via Telegram bot",
    icon: "/upload/notification/telegram.png"
  },
  {
    value: "webhook",
    label: "Webhook",
    description: "Send notifications to custom webhook",
    icon: "/upload/notification/webhook.png"
  },
];

const webhookPayloadTemplates = {
  server: `{
  "alert_type": "server",
  "server_name": "\${server_name}",
  "status": "\${status}",
  "cpu_usage": "\${cpu_usage}",
  "ram_usage": "\${ram_usage}",
  "disk_usage": "\${disk_usage}",
  "network_usage": "\${network_usage}",
  "cpu_temp": "\${cpu_temp}",
  "disk_io": "\${disk_io}",
  "threshold": "\${time}",
  "message": "Server \${server_name} alert: \${status}"
}`,
  service: `{
  "alert_type": "service",
  "service_name": "\${service_name}",
  "status": "\${status}",
  "response_time": "\${response_time}",
  "url": "\${url}",
  "uptime": "\${uptime}",
  "downtime": "\${downtime}",
  "timestamp": "\${time}",
  "message": "Service \${service_name} is \${status}"
}`,
  ssl: `{
  "alert_type": "ssl",
  "domain": "\${domain}",
  "certificate_name": "\${certificate_name}",
  "expiry_date": "\${expiry_date}",
  "days_left": "\${days_left}",
  "issuer": "\${issuer}",
  "serial_number": "\${serial_number}",
  "timestamp": "\${time}",
  "message": "SSL certificate for \${domain} expires in \${days_left} days"
}`
};

const defaultPayloadTemplate = `{
  "alert_type": "general",
  "service_name": "\${service_name}",
  "status": "\${status}",
  "message": "\${message}",
  "timestamp": "\${time}"
}`;

export const NotificationChannelDialog = ({ 
  open, 
  onClose,
  editingConfig 
}: NotificationChannelDialogProps) => {
  const { t } = useLanguage();
  const isEditing = !!editingConfig;
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      notify_name: "",
      notification_type: "telegram" as const,
      enabled: true,
      service_id: "global",
      template_id: "",
    },
  });

  const { watch, reset, setValue } = form;
  const notificationType = watch("notification_type");
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  
  useEffect(() => {
    if (editingConfig) {
      // enabled is already mapped to boolean by alertConfigService.mapRecord
      const enabled = !!editingConfig.enabled;

      reset({
        ...editingConfig,
        enabled
      });
    } else if (open) {
      reset({
        notify_name: "",
        notification_type: "telegram" as const,
        enabled: true,
        service_id: "global",
        template_id: "",
      });
    }
  }, [editingConfig, open, reset]);

  const handleClose = () => {
    onClose(false);
  };

  const insertTemplate = (template: string) => {
    setValue("webhook_payload_template", template);
  };

  const onSubmit = async (values: FormValues) => {
    setIsSubmitting(true);
    try {
      // Handle all notification types including webhook through alert_configurations
      const configData = {
        ...values,
        service_id: values.service_id || "global",
      };
      
      if (isEditing && editingConfig?.id) {
        await alertConfigService.updateAlertConfiguration(editingConfig.id, configData);
      } else {
        await alertConfigService.createAlertConfiguration(configData as any);
      }
      
      onClose(true); // Close with refresh
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save notification channel",
        variant: "destructive"
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={() => handleClose()}>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {isEditing ? t("editChannel") : t("addChannelDialog")}
          </DialogTitle>
        </DialogHeader>
        
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <FormField
              control={form.control}
              name="notify_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("channelName")}</FormLabel>
                  <FormControl>
                    <Input placeholder={t("channelNamePlaceholder")} {...field} />
                  </FormControl>
                  <FormDescription>
                    {t("channelNameDesc")}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="notification_type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("channelType")}</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder={t("selectType")} />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {notificationTypeOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          <div className="flex items-center space-x-3">
                            <img 
                              src={option.icon} 
                              alt={`${option.label} icon`}
                              className="w-5 h-5 object-contain"
                            />
                            <div className="flex flex-col">
                              <span className="font-medium">{t(option.value)}</span>
                              <span className="text-xs text-muted-foreground">{t(option.description)}</span>
                            </div>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {notificationType === "telegram" && (
              <>
                <FormField
                  control={form.control}
                  name="telegram_chat_id"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("telegramChatId")}</FormLabel>
                      <FormControl>
                        <Input placeholder={t("telegramChatIdPlaceholder")} {...field} />
                      </FormControl>
                      <FormDescription>
                        {t("telegramChatIdDesc")}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="bot_token"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("botToken")}</FormLabel>
                      <FormControl>
                        <Input placeholder={t("botTokenPlaceholder")} {...field} type="password" />
                      </FormControl>
                      <FormDescription>
                        {t("botTokenDesc")}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="telegram_thread_id"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Thread ID <span className="text-muted-foreground font-normal">(optional)</span></FormLabel>
                      <FormControl>
                        <Input placeholder="e.g. 12345" {...field} />
                      </FormControl>
                      <FormDescription>
                        Topic/Thread ID for Telegram Supergroup. Leave empty to send to main group chat.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </>
            )}

            {notificationType === "webhook" && (
              <>
                <FormField
                  control={form.control}
                  name="webhook_url"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("webhookUrl")}</FormLabel>
                      <FormControl>
                        <Input placeholder={t("webhookUrlPlaceholder")} {...field} />
                      </FormControl>
                      <FormDescription>
                        {t("webhookUrlDesc")}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="space-y-4">
                  <FormField
                    control={form.control}
                    name="webhook_payload_template"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t("payloadTemplate")}</FormLabel>
                        <FormDescription>
                          {t("payloadTemplateDesc")}
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <Card>
                    <CardHeader className="pb-3">
                      <CardTitle className="text-sm font-medium">{t("payloadTemplates")}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-3">
                        <div className="mt-4">
                          <h4 className="text-sm font-medium mb-2">{t("availablePlaceholders")}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs">
                            <div className="space-y-1">
                              <p className="font-medium text-muted-foreground">{t("server")}</p>
                              <div className="space-y-0.5">
                                <code className="bg-muted px-1 rounded">${'{server_name}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{cpu_usage}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{ram_usage}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{disk_usage}'}</code>
                              </div>
                            </div>
                            <div className="space-y-1">
                              <p className="font-medium text-muted-foreground">{t("service")}</p>
                              <div className="space-y-0.5">
                                <code className="bg-muted px-1 rounded">${'{service_name}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{response_time}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{url}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{uptime}'}</code>
                              </div>
                            </div>
                            <div className="space-y-1">
                              <p className="font-medium text-muted-foreground">{t("ssl")}</p>
                              <div className="space-y-0.5">
                                <code className="bg-muted px-1 rounded">${'{domain}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{expiry_date}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{days_left}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{issuer}'}</code>
                              </div>
                            </div>
                            <div className="space-y-1">
                              <p className="font-medium text-muted-foreground">{t("common")}</p>
                              <div className="space-y-0.5">
                                <code className="bg-muted px-1 rounded">${'{status}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{time}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{message}'}</code><br/>
                                <code className="bg-muted px-1 rounded">${'{threshold}'}</code>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </>
            )}
            
            <FormField
              control={form.control}
              name="enabled"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                  <div className="space-y-0.5">
                    <FormLabel>{t("enabled")}</FormLabel>
                    <FormDescription>
                      {t("enabledDesc")}
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            
            <DialogFooter>
              <Button variant="outline" type="button" onClick={handleClose}>
                {t("cancel")}
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {isEditing ? t("updateChannel") : t("createChannel")}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};