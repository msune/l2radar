/**
 * Deterministic host definitions for the l2radar demo simulation.
 *
 * 15 neighbours total: 9 on eth0, 6 on wlan0.
 *
 * active: true  → last_seen and byte counters update on every write step.
 * active: false → silent host; first_seen === last_seen, never changes.
 *
 * firstOffset: seconds from simulation start when the host first appears.
 *
 * Timeline:
 *   t=0s:  H1, H2, H3, W1, W2       →  5 total
 *   t=4s:  H4                        →  6
 *   t=8s:  W3                        →  7
 *   t=12s: H5, H6                    →  9
 *   t=16s: W4                        → 10
 *   t=20s: H7, H8                    → 12
 *   t=24s: W5, W6                    → 14
 *   t=28s: H9                        → 15  (all hosts visible)
 */

/** Own-interface metadata for each monitored NIC. */
export const ETH0_IF = {
  name: 'eth0',
  mac: 'de:ad:be:ef:00:01',
  ipv4: ['192.168.1.254'],
  ipv6: ['fe80::dead:beff:feef:1'],
}

export const WLAN0_IF = {
  name: 'wlan0',
  mac: 'de:ad:be:ef:ff:01',
  ipv4: ['10.0.0.254'],
  ipv6: ['fe80::dead:beff:feff:1'],
}

/**
 * Per-host deterministic TX increment (bytes per 2-second write step).
 * Keyed by host id.
 */
export const TX_PER_STEP = {
  H1: 12288,  // Dell server — moderate traffic
  H2: 4096,   // Raspberry Pi — light
  H4: 81920,  // TP-Link router — heavy (gateway traffic)
  H5: 2048,   // VMware VM — light
  H7: 8192,   // Samsung TV — streaming
  H8: 20480,  // Synology NAS — storage traffic
  W1: 65536,  // TP-Link AP — all wlan0 traffic passes through
  W2: 16384,  // Intel laptop — active user
  W4: 4096,   // Samsung phone — moderate
  W6: 1024,   // Siemens IoT — low-rate sensor data
}

/** RX is roughly 2× TX for most devices. */
export const RX_PER_STEP = {
  H1: 24576,
  H2: 8192,
  H4: 163840,
  H5: 4096,
  H7: 16384,
  H8: 40960,
  W1: 131072,
  W2: 32768,
  W4: 8192,
  W6: 2048,
}

/** All 15 neighbours. */
export const HOSTS = [
  // ── eth0 ──────────────────────────────────────────────────────────────────
  {
    id: 'H1', iface: 'eth0', active: true, firstOffset: 0,
    mac: 'b8:ca:3a:a1:02:03',          // Dell Inc.
    ipv4: ['192.168.1.10'],
    ipv6: ['fe80::baca:3aff:fea1:203'],
  },
  {
    id: 'H2', iface: 'eth0', active: true, firstOffset: 0,
    mac: 'dc:a6:32:11:22:33',          // Raspberry Pi Trading Ltd
    ipv4: ['192.168.1.20'],
    ipv6: ['fe80::dea6:32ff:fe11:2233'],
  },
  {
    id: 'H3', iface: 'eth0', active: false, firstOffset: 0,
    mac: 'a0:78:17:aa:bb:cc',          // Apple, Inc. (iPhone — passed through)
    ipv4: ['192.168.1.30'],
    ipv6: [],
  },
  {
    id: 'H4', iface: 'eth0', active: true, firstOffset: 4,
    mac: '50:c7:bf:dd:ee:ff',          // TP-LINK TECHNOLOGIES (gateway)
    ipv4: ['192.168.1.1', '10.0.0.1'],
    ipv6: ['fe80::52c7:bfff:fedd:eeff'],
  },
  {
    id: 'H5', iface: 'eth0', active: true, firstOffset: 12,
    mac: '00:0c:29:11:33:55',          // VMware, Inc. (virtual machine)
    ipv4: ['192.168.1.50'],
    ipv6: ['fe80::20c:29ff:fe11:3355'],
  },
  {
    id: 'H6', iface: 'eth0', active: false, firstOffset: 12,
    mac: 'fc:3f:db:77:88:99',          // Hewlett Packard (network printer)
    ipv4: ['192.168.1.60'],
    ipv6: [],
  },
  {
    id: 'H7', iface: 'eth0', active: true, firstOffset: 20,
    mac: '94:8b:c1:ab:cd:ef',          // Samsung Electronics (smart TV)
    ipv4: ['192.168.1.70'],
    ipv6: ['fe80::968b:c1ff:feab:cdef'],
  },
  {
    id: 'H8', iface: 'eth0', active: true, firstOffset: 20,
    mac: '00:11:32:44:55:66',          // Synology Incorporated (NAS)
    ipv4: ['192.168.1.80', '192.168.2.80'],
    ipv6: ['fe80::211:32ff:fe44:5566'],
  },
  {
    id: 'H9', iface: 'eth0', active: false, firstOffset: 28,
    mac: 'cc:46:d6:01:02:03',          // Cisco Systems (switch — silent)
    ipv4: ['192.168.1.90'],
    ipv6: [],
  },

  // ── wlan0 ─────────────────────────────────────────────────────────────────
  {
    id: 'W1', iface: 'wlan0', active: true, firstOffset: 0,
    mac: '50:c7:bf:01:02:03',          // TP-LINK (access point)
    ipv4: ['10.0.0.1', '192.168.2.1'],
    ipv6: ['fe80::52c7:bfff:fe01:203'],
  },
  {
    id: 'W2', iface: 'wlan0', active: true, firstOffset: 0,
    mac: '10:02:b5:aa:bb:cc',          // Intel Corporate (laptop)
    ipv4: ['10.0.0.10'],
    ipv6: ['fe80::1202:b5ff:feaa:bbcc'],
  },
  {
    id: 'W3', iface: 'wlan0', active: false, firstOffset: 8,
    mac: '18:65:90:11:22:33',          // Apple, Inc. (iPhone — guest)
    ipv4: ['10.0.0.20'],
    ipv6: [],
  },
  {
    id: 'W4', iface: 'wlan0', active: true, firstOffset: 16,
    mac: '60:6b:bd:00:11:22',          // Samsung Electronics (Android phone)
    ipv4: ['10.0.0.30'],
    ipv6: ['fe80::626b:bdff:fe00:1122'],
  },
  {
    id: 'W5', iface: 'wlan0', active: false, firstOffset: 24,
    mac: '3c:28:6d:cc:dd:ee',          // Google, Inc. (Pixel — guest)
    ipv4: ['10.0.0.40'],
    ipv6: [],
  },
  {
    id: 'W6', iface: 'wlan0', active: true, firstOffset: 24,
    mac: 'd4:f5:27:55:66:77',          // SIEMENS AG (IoT sensor)
    ipv4: ['10.0.0.50'],
    ipv6: ['fe80::d6f5:27ff:fe55:6677'],
  },
]
