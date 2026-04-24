/// <reference path="../pb_data/types.d.ts" />
// Updates default notification templates to use HTML bold formatting
// (parse_mode: HTML is set on all Telegram messages)
migrate((app) => {
  const upsertRecord = (collection, filterName, data) => {
    let records;
    try {
      records = app.findRecordsByFilter(collection.id, `name = "${filterName}"`, "", 1, 0);
    } catch(e) {
      records = [];
    }

    if (records && records.length > 0) {
      const rec = records[0];
      Object.keys(data).forEach(k => rec.set(k, data[k]));
      app.save(rec);
    } else {
      const rec = new Record(collection);
      Object.keys(data).forEach(k => rec.set(k, data[k]));
      app.save(rec);
    }
  };

  // ── Service / Uptime ────────────────────────────────────────────────────────
  const serviceCol = app.findCollectionByNameOrId("service_notification_templates");

  upsertRecord(serviceCol, "Default Uptime Template", {
    name: "Default Uptime Template",
    up_message: "🟢 <b>Service UP</b>\n\n📡 <b>Service:</b> ${service_name}\n🔗 <b>URL:</b> ${url}\n⚡ <b>Response:</b> ${response_time}\n🌍 <b>Type:</b> ${service_type}\n🕐 <b>Time:</b> ${time}",
    down_message: "🔴 <b>Service DOWN</b>\n\n📡 <b>Service:</b> ${service_name}\n🔗 <b>URL:</b> ${url}\n❌ <b>Error:</b> ${error_message}\n🌍 <b>Type:</b> ${service_type}\n🕐 <b>Time:</b> ${time}",
    warning_message: "⚠️ <b>Service WARNING</b>\n\n📡 <b>Service:</b> ${service_name}\n🔗 <b>URL:</b> ${url}\n⚡ <b>Response:</b> ${response_time}\n🕐 <b>Time:</b> ${time}",
    maintenance_message: "🔧 <b>Service MAINTENANCE</b>\n\n📡 <b>Service:</b> ${service_name}\n🕐 <b>Time:</b> ${time}",
    incident_message: "🚨 <b>INCIDENT Detected</b>\n\n📡 <b>Service:</b> ${service_name}\n🔗 <b>URL:</b> ${url}\n❌ <b>Error:</b> ${error_message}\n🕐 <b>Time:</b> ${time}",
    resolved_message: "✅ <b>Incident RESOLVED</b>\n\n📡 <b>Service:</b> ${service_name}\n🔗 <b>URL:</b> ${url}\n🕐 <b>Time:</b> ${time}",
    placeholder: "${service_name}, ${status}, ${url}, ${response_time}, ${error_message}, ${service_type}, ${time}",
  });

  // ── Server monitoring ───────────────────────────────────────────────────────
  const serverCol = app.findCollectionByNameOrId("server_notification_templates");

  upsertRecord(serverCol, "Default Server Template", {
    name: "Default Server Template",
    up_message: "🟢 <b>Server UP</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🕐 <b>Time:</b> ${time}",
    down_message: "🔴 <b>Server DOWN</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🕐 <b>Time:</b> ${time}",
    warning_message: "⚠️ <b>Server WARNING</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🕐 <b>Time:</b> ${time}",
    paused_message: "⏸️ <b>Server PAUSED</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🕐 <b>Time:</b> ${time}",
    cpu_message: "🖥️ <b>High CPU Alert</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>CPU Usage:</b> ${cpu_usage}\n⚠️ <b>Threshold:</b> ${threshold}\n🕐 <b>Time:</b> ${time}",
    ram_message: "💾 <b>High RAM Alert</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>RAM Usage:</b> ${ram_usage}\n⚠️ <b>Threshold:</b> ${threshold}\n🕐 <b>Time:</b> ${time}",
    disk_message: "💿 <b>High Disk Alert</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>Disk Usage:</b> ${disk_usage}\n⚠️ <b>Threshold:</b> ${threshold}\n🕐 <b>Time:</b> ${time}",
    network_message: "🌐 <b>High Network Alert</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>Network:</b> ${network_usage}\n🕐 <b>Time:</b> ${time}",
    cpu_temp_message: "🌡️ <b>High CPU Temperature</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🌡️ <b>Temperature:</b> ${cpu_temp}\n⚠️ <b>Threshold:</b> ${threshold}\n🕐 <b>Time:</b> ${time}",
    disk_io_message: "💽 <b>High Disk I/O Alert</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>Disk I/O:</b> ${disk_io}\n🕐 <b>Time:</b> ${time}",
    restore_cpu_message: "✅ <b>CPU Usage Restored</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>CPU Usage:</b> ${cpu_usage}\n🕐 <b>Time:</b> ${time}",
    restore_ram_message: "✅ <b>RAM Usage Restored</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>RAM Usage:</b> ${ram_usage}\n🕐 <b>Time:</b> ${time}",
    restore_disk_message: "✅ <b>Disk Usage Restored</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>Disk Usage:</b> ${disk_usage}\n🕐 <b>Time:</b> ${time}",
    restore_network_message: "✅ <b>Network Usage Restored</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🕐 <b>Time:</b> ${time}",
    restore_cpu_temp_message: "✅ <b>CPU Temperature Restored</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n🌡️ <b>Temperature:</b> ${cpu_temp}\n🕐 <b>Time:</b> ${time}",
    restore_disk_io_message: "✅ <b>Disk I/O Restored</b>\n\n🖥️ <b>Server:</b> ${service_name}\n📍 <b>Host:</b> ${hostname}\n📊 <b>Disk I/O:</b> ${disk_io}\n🕐 <b>Time:</b> ${time}",
    placeholder: "${service_name}, ${hostname}, ${status}, ${cpu_usage}, ${ram_usage}, ${disk_usage}, ${network_usage}, ${cpu_temp}, ${disk_io}, ${threshold}, ${time}",
  });

}, (app) => {
  // Rollback: restore plain text format
  const upsertRecord = (collection, filterName, data) => {
    let records;
    try {
      records = app.findRecordsByFilter(collection.id, `name = "${filterName}"`, "", 1, 0);
    } catch(e) { records = []; }
    if (records && records.length > 0) {
      const rec = records[0];
      Object.keys(data).forEach(k => rec.set(k, data[k]));
      app.save(rec);
    }
  };

  const serviceCol = app.findCollectionByNameOrId("service_notification_templates");
  upsertRecord(serviceCol, "Default Uptime Template", {
    up_message: "✅ Service ${service_name} is UP\n - URL: ${url}\n - Response: ${response_time}\n - Time: ${time}",
    down_message: "🔴 Service ${service_name} is DOWN\n - URL: ${url}\n - Error: ${error_message}\n - Time: ${time}",
    warning_message: "⚠️ Service ${service_name} WARNING\n - URL: ${url}\n - Time: ${time}",
    maintenance_message: "🔧 Service ${service_name} MAINTENANCE\n - Time: ${time}",
    incident_message: "🚨 INCIDENT on ${service_name}\n - URL: ${url}\n - Time: ${time}",
    resolved_message: "✅ Incident RESOLVED for ${service_name}\n - Time: ${time}",
  });

  const serverCol = app.findCollectionByNameOrId("server_notification_templates");
  upsertRecord(serverCol, "Default Server Template", {
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
  });
});
