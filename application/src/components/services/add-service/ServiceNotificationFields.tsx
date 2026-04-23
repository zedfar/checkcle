import { FormControl, FormField, FormItem, FormLabel, FormDescription } from "@/components/ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";
import { UseFormReturn } from "react-hook-form";
import { ServiceFormData } from "./types";
import { useQuery } from "@tanstack/react-query";
import { alertConfigService, AlertConfiguration } from "@/services/alertConfigService";
import { serviceNotificationTemplateService, ServiceNotificationTemplate } from "@/services/serviceNotificationTemplateService";
import { useState, useEffect } from "react";
import { useLanguage } from "@/contexts/LanguageContext.tsx";

interface ServiceNotificationFieldsProps {
  form: UseFormReturn<ServiceFormData>;
}

export function ServiceNotificationFields({ form }: ServiceNotificationFieldsProps) {
	const { t } = useLanguage();
  const [alertConfigs, setAlertConfigs] = useState<AlertConfiguration[]>([]);
  
  // Get the current form values for debugging
  const notificationStatus = form.watch("notificationStatus");
  const notificationChannels = form.watch("notificationChannels") || [];
  const alertTemplate = form.watch("alertTemplate");
   
  // Fetch alert configurations for notification channels
  const { data: alertConfigsData } = useQuery({
    queryKey: ['alertConfigs'],
    queryFn: () => alertConfigService.getAlertConfigurations(),
  });

  // Fetch service notification templates
  const { data: serviceTemplates, isLoading: isLoadingTemplates } = useQuery({
    queryKey: ['serviceNotificationTemplates'],
    queryFn: () => serviceNotificationTemplateService.getTemplates(),
  });
  
  // Update alert configs when data is loaded
  useEffect(() => {
    if (alertConfigsData) {
      // Only show enabled channels (enabled is mapped to boolean by alertConfigService)
      const enabledChannels = alertConfigsData.filter(config => config.enabled === true);
      setAlertConfigs(enabledChannels);
      
      // Debug log to check what alert configs are loaded
    }
  }, [alertConfigsData]);

  // Debug log for service templates
  useEffect(() => {
    if (serviceTemplates) {
     // console.log("Loaded service notification templates:", serviceTemplates);
    }
  }, [serviceTemplates]);

  // Log when form values change to debug
  useEffect(() => {
   // console.log("Notification values changed:", {
   //   notificationStatus: form.getValues("notificationStatus"),
   //   notificationChannels: form.getValues("notificationChannels")
   // });
  }, [form.watch("notificationStatus"), form.watch("notificationChannels")]);

  const handleChannelAdd = (channelId: string) => {
    const currentChannels = form.getValues("notificationChannels") || [];
    if (!currentChannels.includes(channelId)) {
      form.setValue("notificationChannels", [...currentChannels, channelId]);
    }
  };

  const handleChannelRemove = (channelId: string) => {
    const currentChannels = form.getValues("notificationChannels") || [];
    form.setValue("notificationChannels", currentChannels.filter(id => id !== channelId));
  };

  const getSelectedChannelNames = () => {
    return (notificationChannels || []).map(channelId => {
      const config = alertConfigs.find(c => c.id === channelId);
      return config ? `${config.notify_name} (${config.notification_type})` : channelId;
    });
  };
  
  return (
    <>
      <FormField
        control={form.control}
        name="notificationStatus"
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <FormLabel className="text-base">
	              {t("enableNotifications")}
              </FormLabel>
              <FormDescription>
	              {t("enableNotificationsDesc")}
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={field.value === "enabled"}
                onCheckedChange={(checked) => {
                  field.onChange(checked ? "enabled" : "disabled");
                  // Clear notification channels when disabled
                  if (!checked) {
                    form.setValue("notificationChannels", []);
                  }
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      
      <FormField
        control={form.control}
        name="notificationChannels"
        render={({ field }) => (
          <FormItem>
            <FormLabel>{t("notificationChannels")}</FormLabel>
            <FormDescription>
              {notificationStatus === "enabled" 
                ? t("notificationChannelsEnabledDesc")
                : t("notificationChannelsDesc")}
            </FormDescription>
            
            {/* Display selected channels as badges */}
            {notificationChannels && notificationChannels.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-2">
                {getSelectedChannelNames().map((channelName, index) => (
                  <Badge key={notificationChannels[index]} variant="secondary" className="flex items-center gap-1">
                    {channelName}
                    <X 
                      className="h-3 w-3 cursor-pointer" 
                      onClick={() => handleChannelRemove(notificationChannels[index])}
                    />
                  </Badge>
                ))}
              </div>
            )}
            
            <FormControl>
              <Select 
                onValueChange={handleChannelAdd}
                disabled={notificationStatus !== "enabled"}
                value="" // Always reset to empty after selection
              >
                <SelectTrigger className={notificationStatus !== "enabled" ? 'opacity-50' : ''}>
                  <SelectValue placeholder={t("notificationChannelsPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {alertConfigs
                    .filter(config => !notificationChannels?.includes(config.id || ""))
                    .map((config) => (
                      <SelectItem key={config.id} value={config.id || ""}>
                        {config.notify_name} ({config.notification_type})
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </FormControl>
          </FormItem>
        )}
      />
            
      <FormField
        control={form.control}
        name="alertTemplate"
        render={({ field }) => {
         // console.log("Rendering alert template field with value:", field.value);
          
          return (
            <FormItem>
              <FormLabel>{t("alertTemplate")}</FormLabel>
              <FormControl>
                <Select 
                  onValueChange={(value) => {
                    field.onChange(value);
                  }} 
                  value={field.value || ""}
                  disabled={notificationStatus !== "enabled" || isLoadingTemplates}
                >
                  <SelectTrigger className={notificationStatus !== "enabled" ? 'opacity-50' : ''}>
                    <SelectValue placeholder={isLoadingTemplates ? t("alertTemplateLoading") : t("alertTemplatePlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {serviceTemplates?.map((template) => (
                      <SelectItem key={template.id} value={template.id}>
                        {template.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormControl>
              <FormDescription>
                {notificationStatus === "enabled"
                  ? t("alertTemplateEnabledDesc")
                  : t("alertTemplateDesc")}
              </FormDescription>
            </FormItem>
          );
        }}
      />
    </>
  );
}