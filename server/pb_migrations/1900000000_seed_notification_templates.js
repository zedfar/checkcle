/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const upsertRecord = (collection, filterName, data) => {
    let records;
    try {
      records = app.findRecordsByFilter(collection.id, `name = "${filterName}"`, "", 1, 0);
    } catch(e) {
      records = [];
    }

    if (records && records.length > 0) {
      // Already exists — update in case fields changed
      const rec = records[0];
      Object.keys(data).forEach(k => rec.set(k, data[k]));
      app.save(rec);
    } else {
      // Fresh — create
      const rec = new Record(collection);
      Object.keys(data).forEach(k => rec.set(k, data[k]));
      app.save(rec);
    }
  };

  // ── Service / Uptime notification templates ─────────────────────────────────
  const serviceCol = app.findCollectionByNameOrId("service_notification_templates");

  upsertRecord(serviceCol, "Default Uptime Template", {
    name: "Default Uptime Template",
    up_message: "✅ Service ${service_name} is UP\n - URL: ${url}\n - Response: ${response_time}\n - Time: ${time}",
    down_message: "🔴 Service ${service_name} is DOWN\n - URL: ${url}\n - Error: ${error_message}\n - Time: ${time}",
    warning_message: "⚠️ Service ${service_name} WARNING\n - URL: ${url}\n - Time: ${time}",
    maintenance_message: "🔧 Service ${service_name} MAINTENANCE\n - Time: ${time}",
    incident_message: "🚨 INCIDENT on ${service_name}\n - URL: ${url}\n - Time: ${time}",
    resolved_message: "✅ Incident RESOLVED for ${service_name}\n - Time: ${time}",
    placeholder: "${service_name}, ${status}, ${url}, ${response_time}, ${error_message}, ${time}",
  });

  // ── Server monitoring notification templates ────────────────────────────────
  const serverCol = app.findCollectionByNameOrId("server_notification_templates");

  upsertRecord(serverCol, "Default Server Template", {
    name: "Default Server Template",
    up_message: "✅ Server ${service_name} is UP\n - Host: ${hostname}\n - Time: ${time}",
    down_message: "🔴 Server ${service_name} is DOWN\n - Host: ${hostname}\n - Time: ${time}",
    warning_message: "⚠️ Server ${service_name} WARNING\n - Host: ${hostname}\n - Time: ${time}",
    paused_message: "⏸️ Server ${service_name} PAUSED\n - Time: ${time}",
    cpu_message: "🖥️ High CPU on ${service_name}: ${cpu_usage} (threshold: ${threshold})\n - Host: ${hostname}\n - Time: ${time}",
    ram_message: "💾 High RAM on ${service_name}: ${ram_usage} (threshold: ${threshold})\n - Host: ${hostname}\n - Time: ${time}",
    disk_message: "💿 High Disk on ${service_name}: ${disk_usage} (threshold: ${threshold})\n - Host: ${hostname}\n - Time: ${time}",
    network_message: "🌐 High Network on ${service_name}: ${network_usage}\n - Host: ${hostname}\n - Time: ${time}",
    cpu_temp_message: "🌡️ High CPU Temp on ${service_name}: ${cpu_temp}\n - Host: ${hostname}\n - Time: ${time}",
    disk_io_message: "💽 High Disk I/O on ${service_name}: ${disk_io}\n - Host: ${hostname}\n - Time: ${time}",
    restore_cpu_message: "✅ CPU restored on ${service_name}: ${cpu_usage}\n - Host: ${hostname}\n - Time: ${time}",
    restore_ram_message: "✅ RAM restored on ${service_name}: ${ram_usage}\n - Host: ${hostname}\n - Time: ${time}",
    restore_disk_message: "✅ Disk restored on ${service_name}: ${disk_usage}\n - Host: ${hostname}\n - Time: ${time}",
    restore_network_message: "✅ Network restored on ${service_name}\n - Host: ${hostname}\n - Time: ${time}",
    restore_cpu_temp_message: "✅ CPU Temp restored on ${service_name}\n - Host: ${hostname}\n - Time: ${time}",
    restore_disk_io_message: "✅ Disk I/O restored on ${service_name}\n - Host: ${hostname}\n - Time: ${time}",
    placeholder: "${service_name}, ${hostname}, ${status}, ${cpu_usage}, ${ram_usage}, ${disk_usage}, ${threshold}, ${time}",
  });

}, (app) => {
  // Rollback: hapus record seed berdasarkan nama
  const tryDelete = (collectionName, name) => {
    try {
      const col = app.findCollectionByNameOrId(collectionName);
      const recs = app.findRecordsByFilter(col.id, `name = "${name}"`, "", 1, 0);
      if (recs && recs.length > 0) app.delete(recs[0]);
    } catch(e) {}
  };

  tryDelete("service_notification_templates", "Default Uptime Template");
  tryDelete("server_notification_templates", "Default Server Template");
});
