plugin_paths = { "/usr/share/jitsi-meet/prosody-plugins/" }

-- domain mapper options, must at least have domain base set to use the mapper
muc_mapper_domain_base = "{{ jitsi_fqdn }}";

turncredentials_secret = "9fu59Z6GcVfWaOoH";

turncredentials = {
  { type = "stun", host = "{{ jitsi_fqdn }}", port = "3478" },
  { type = "turn", host = "{{ jitsi_fqdn }}", port = "3478", transport = "udp" },
  {% if https %}
  { type = "turns", host = "{{ jitsi_fqdn }}", port = "443", transport = "tcp" }
  {% endif %}
};

cross_domain_bosh = false;
consider_bosh_secure = true;
-- https_ports = { }; -- Remove this line to prevent listening on port 5284

-- https://ssl-config.mozilla.org/#server=haproxy&version=2.1&config=intermediate&openssl=1.1.0g&guideline=5.4
ssl = {
  protocol = "tlsv1_2+";
  ciphers = "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"
}

VirtualHost "{{ jitsi_fqdn }}"
        -- enabled = false -- Remove this line to enable this host
        authentication = "anonymous"
        -- Properties below are modified by jitsi-meet-tokens package config
        -- and authentication above is switched to "token"
        -- app_id="mattermost"
        -- app_secret="jitsi_jwt_secret"
        -- allow_empty_token=false
        -- Assign this host a certificate for TLS, otherwise it would use the one
        -- set in the global section (if any).
        -- Note that old-style SSL on port 5223 only supports one certificate, and will always
        -- use the global one.
        ssl = {
                key = "/etc/prosody/certs/{{ jitsi_fqdn }}.key";
                certificate = "/etc/prosody/certs/{{ jitsi_fqdn }}.crt";
        }
        speakerstats_component = "speakerstats.{{ jitsi_fqdn }}"
        conference_duration_component = "conferenceduration.{{ jitsi_fqdn }}"
        -- we need bosh
        modules_enabled = {
            "bosh";
            "pubsub";
            "ping"; -- Enable mod_ping
            "speakerstats";
            "turncredentials";
            "conference_duration";
            "muc_lobby_rooms";
            "presence_identity";
        }
        c2s_require_encryption = false
        lobby_muc = "lobby.{{ jitsi_fqdn }}"
        main_muc = "conference.{{ jitsi_fqdn }}"
        -- muc_lobby_whitelist = { "recorder.{{ jitsi_fqdn }}" } -- Here we can whitelist jibri to enter lobby enabled rooms

Component "conference.{{ jitsi_fqdn }}" "muc"
    storage = "memory"
    modules_enabled = {
        "muc_meeting_id";
    }
    admins = { "focus@auth.{{ jitsi_fqdn }}" }
    muc_room_locking = false
    muc_room_default_public_jids = true

-- internal muc component
Component "internal.auth.{{ jitsi_fqdn }}" "muc"
    storage = "memory"
    modules_enabled = {
      "ping";
    }
    admins = { "focus@auth.{{ jitsi_fqdn }}", "jvb@auth.{{ jitsi_fqdn }}" }
    muc_room_locking = false
    muc_room_default_public_jids = true

VirtualHost "auth.{{ jitsi_fqdn }}"
    ssl = {
        key = "/etc/prosody/certs/auth.{{ jitsi_fqdn }}.key";
        certificate = "/etc/prosody/certs/auth.{{ jitsi_fqdn }}.crt";
    }
    authentication = "internal_plain"

Component "focus.{{ jitsi_fqdn }}"
    component_secret = "{{ jitsi_focus_secret }}"

Component "jitsi-videobridge.{{ jitsi_fqdn }}"
    component_secret = "{{ jitsi_jvb_secret }}"

Component "speakerstats.{{ jitsi_fqdn }}" "speakerstats_component"
    muc_component = "conference.{{ jitsi_fqdn }}"

Component "conferenceduration.{{ jitsi_fqdn }}" "conference_duration_component"
    muc_component = "conference.{{ jitsi_fqdn }}"

Component "lobby.{{ jitsi_fqdn }}" "muc"
    storage = "memory"
    restrict_room_creation = true
    muc_room_locking = false
    muc_room_default_public_jids = true
