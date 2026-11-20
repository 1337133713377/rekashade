import { computed, ref } from 'vue'

type Dict = { [k: string]: string | Dict }

const en: Dict = {
  common: {
    on: 'ON', off: 'OFF', save: 'Save', cancel: 'Cancel', apply: 'Apply', delete: 'Delete', duplicate: 'Duplicate',
    rename: 'Rename', copy: 'Copy', copied: 'Copied!', paste: 'Paste', import: 'Import', export: 'Export', reset: 'Reset',
    close: 'Close', add: 'Add', remove: 'Remove', custom: 'Custom', none: 'None', loading: 'Loading…', error: 'Error',
    display: 'Display {n}', allDisplays: 'All displays', active: 'Active', idle: 'Idle', open: 'Open', confirm: 'Confirm',
    minimize: 'Minimize', closeWindow: 'Close', enabled: 'Enabled', default: 'Default',
  },
  nav: {
    menu: 'MENU', general: 'GENERAL', display: 'Display', presets: 'Presets', resolution: 'Resolution', crosshair: 'Crosshair',
    reticle: 'Reticle', binds: 'Binds', optimizer: 'Optimizer', theme: 'Theme', community: 'Community', account: 'Account',
    settings: 'Settings', about: 'About', legal: 'Legal', tuning: 'TUNING', customLook: 'Custom look', tuningOff: 'Tuning off',
    statusActive: 'Active', statusIdle: 'Idle', update: 'Update {v}',
  },
  display: {
    title: 'Display Tuning', onDisplay: '{preset} on {display}', off: 'Tuning is off — your displays look exactly as before.',
    adjust: 'Adjust', tone: 'Tone', color: 'Color', dblclick: 'Double-click a slider to reset it.',
    hue: 'Hue', saturation: 'Saturation', vibrance: 'Vibrance', red: 'Red', green: 'Green', blue: 'Blue',
    brightness: 'Brightness', contrast: 'Contrast', gamma: 'Gamma', temperature: 'Temperature',
    allScreens: 'All screens', gpu: 'GPU', windowsEngine: 'Windows', startWithWindows: 'Start with Windows',
    noteSat: 'Hue and saturation apply to every screen at once.', resetNeutral: 'Reset to neutral',
    preview: 'Preview', before: 'BEFORE', after: 'AFTER', demo: 'Demo', curve: 'Curve',
    dragHint: 'Drag to compare before and after.', curveHint: 'Output of the red, green and blue channels for every input level.',
    colorPresets: 'Color presets', allPresets: 'All presets', canyon: 'Canyon', neon: 'Neon City', forest: 'Forest',
    limited: 'Windows limited this look to {pct}% on {display}.', unlock: 'Unlock full range',
    gammaError: '{display} refuses gamma changes (is HDR on?). Tone sliders have no effect there.',
    matrixError: 'Windows colour engine unavailable: {error}',
    vibGpu: 'Vibrance uses your {gpu} driver: zero latency, works even in exclusive fullscreen.',
    vibMatrix: 'This GPU has no driver vibrance, so Reka Shade uses the Windows colour engine for it.',
    rule: 'Game profile active: {exe}', welcomeTitle: 'Welcome to Reka Shade',
    welcomeText: 'Welcome! Pick a preset or move the sliders, then flip the switch. Nothing here touches your games.',
  },
  presets: {
    title: 'Presets', subtitle: 'Looks you can apply in one click — and share with a code.', saveCurrent: 'Save current look',
    importCode: 'Import code', builtin: 'Built-in', mine: 'My presets', applyTo: 'Apply to {display}', active: 'Active',
    copyCode: 'Copy code', duplicate: 'Duplicate', rename: 'Rename', delete: 'Delete', pin: 'Pin to Display page',
    unpin: 'Unpin', namePrompt: 'Preset name', newName: 'My preset {n}', saved: 'Preset “{name}” saved.',
    rules: 'Game profiles', rulesHint: 'Switch preset automatically while a game is in focus. Reka Shade only reads the window list — it never opens the game.',
    addRule: 'Add game', pickRunning: 'Running apps', typeExe: 'or type an .exe name', noRules: 'No game profiles yet.',
    rulesNeedTuning: 'Game profiles work while Display tuning is on.', emptyMine: 'Save your current look to create your first preset.',
    pasteCode: 'Paste a preset code (RKP1-…)', invalidCode: 'That is not a valid preset code.', imported: 'Preset “{name}” imported.',
    deleteConfirm: 'Delete “{name}”?',
    tags: { neutral: 'Neutral', vivid: 'Vivid', warm: 'Warm', cool: 'Cool', cinematic: 'Cinematic', competitive: 'Competitive', dark: 'Dark scenes', mono: 'Mono', custom: 'Custom', shared: 'Shared' },
  },
  res: {
    title: 'Stretched Resolution', subtitle: 'Play in 4:3 or any size, stretched to fill your screen. Only the Windows desktop mode changes — the game is never touched.',
    native: 'Native {w}×{h}', stretched: 'Stretched {w}×{h}', choose: 'Choose a resolution', display: 'Display', custom: 'Custom',
    width: 'Width', height: 'Height', refresh: 'Refresh rate', auto: 'Best available', scaling: 'Scaling', stretch: 'Stretched',
    aspect: 'Black bars', center: 'Centered', notListed: 'Not in Windows’ list for this display. Create it as a custom resolution in the NVIDIA / AMD / Intel control panel first.',
    test: 'Test this resolution', revert: 'Back to native', keepTitle: 'Keep this resolution?', keepBody: 'Going back automatically in {s} s unless you confirm.',
    keep: 'Keep it', revertNow: 'Revert now', preview: 'How it fills your screen', panel: 'Your {w}×{h} screen', game: 'Game at {w}×{h}',
    autoTitle: 'Auto stretch', autoHint: 'Switch to the stretched resolution while one of these games runs and back to native when it closes.',
    revertOnExit: 'Restore native resolution when Reka Shade closes', needConfirm: 'Test and keep the resolution once to unlock auto stretch and the hotkey.',
    hotkey: 'Toggle anytime with {key}', popular: 'Popular', addGame: 'Add game', noGames: 'No games yet.', confirmed: 'Ready',
    current: 'Current: {w}×{h} @ {hz} Hz', inGame: 'Tip: set the same resolution in the game, in fullscreen or borderless.',
  },
  xh: {
    title: 'Crosshair', subtitle: 'Stays on top of your games and apps.', designs: 'Designs', color: 'COLOR', outlineColor: 'OUTLINE',
    my: 'MY CROSSHAIRS', saveCurrent: 'Save current', copyCode: 'Copy code', pasteCode: 'Paste code',
    codesHint: 'Share codes start with RKX1. Copy one, send it to a friend, and they press Paste code.', customize: 'Customize',
    actualSize: 'ACTUAL SIZE', ring: 'Ring', arm: 'Arm length', gap: 'Gap', thickness: 'Thickness', dot: 'Dot', outline: 'Outline',
    opacity: 'Opacity', rotate: 'Rotate 45°', tStyle: 'T-shape', slotHint: 'Click to load · right-click to clear', empty: 'Empty',
    pastePrompt: 'Paste a crosshair code (RKX1-…)', invalid: 'That is not a valid crosshair code.', savedSlot: 'Saved to slot {n}.',
    slotsFull: 'All slots are full — right-click one to clear it.', loaded: 'Crosshair loaded.', pickColor: 'Pick a colour',
    reticleActive: 'The image reticle is the active overlay. Turn this on to switch back to the crosshair.',
  },
  ret: {
    title: 'Reticle', subtitle: 'Image reticles, plus where and when the overlay shows up.', library: 'Library', upload: 'Upload image',
    uploadHint: 'PNG or SVG with transparency. Resized to 256 px max.', size: 'Size', color: 'Color', rotation: 'Rotation', glow: 'Glow',
    opacity: 'Opacity', yours: 'Your image', placement: 'Placement', screen: 'Screen', offset: 'Fine offset', offsetX: 'Horizontal',
    offsetY: 'Vertical', center: 'Re-center', follow: 'Follow the active window', followHint: 'Centers on windowed games instead of the whole screen.',
    onlyFs: 'Only over fullscreen apps', onlyFsHint: 'Hidden on the desktop and in normal windows.', capture: 'Hide from screenshots & streams',
    captureHint: 'OBS, Discord and recordings won’t see it (Windows 10 2004+).', hold: 'Hide while holding', holdNone: 'No key — set one in Binds',
    reasonOff: 'Overlay is off', reasonHold: 'Hidden while the key is held', reasonFs: 'Waiting for a fullscreen app', reasonImg: 'Rendering…',
    visible: 'Showing on screen', behavior: 'Behavior',
    names: { holo: 'Holo', chevron: 'Chevron', 'mil-dot': 'Mil-dot', triangle: 'Triangle', bracket: 'Brackets', diamond: 'Diamond', 'tactical-x': 'Tactical X', 'circle-split': 'Split ring' },
  },
  binds: {
    title: 'Binds', subtitle: 'Global hotkeys that work in game. Keys are only read, never blocked — your game still receives them.',
    toggleTuning: 'Toggle display tuning', toggleCrosshair: 'Toggle crosshair', nextPreset: 'Next preset', prevPreset: 'Previous preset',
    toggleStretch: 'Toggle stretched resolution', showApp: 'Open Reka Shade', panic: 'Panic — turn everything off', holdHide: 'Hide crosshair while held',
    press: 'Press a key…', unbound: 'Not set', clear: 'Clear', change: 'Change', reset: 'Restore defaults',
    holdHint: 'Tip: bind Mouse 2 (right click) to hide the crosshair while aiming down sights.', conflict: 'Already used by “{action}”.',
    mouseHint: 'Mouse buttons 2–5 work too. Esc cancels.', how: 'How binds work',
    howText: 'Reka Shade checks the keyboard state about 60 times a second, like any background app. No keyboard hook, no injected input, nothing in the game process.',
    hold: 'HOLD', toggle: 'PRESS', tip: 'Tip',
  },
  opt: {
    title: 'Optimizer', subtitle: 'Real Windows tweaks for smoother games. Undo puts it all back.', undoAll: 'Undo all', optimizeAll: 'Optimize all',
    clean: 'Clean temp files', cleanDesc: 'Deletes files older than a day from your own Temp folder. Files in use are skipped.', cleanBtn: 'Clean',
    cleaned: 'Removed {files} files ({size}). {skipped} files in use were skipped.', alreadyOn: 'Already on', alreadyOnHint: 'Windows already had this set — nothing to undo.',
    win11: 'Windows 11 only', failed: 'Couldn’t change “{name}”: {error}', allDone: 'All tweaks applied.', allUndone: 'Everything is back the way it was.',
    t: {
      powerPlan: ['High performance power plan', 'CPU stays at full speed instead of saving power. Steadier frame times.'],
      gameMode: ['Game Mode', 'Windows gives the game you play first claim on CPU and GPU time.'],
      gameDVR: ['Background recording off', 'Stops Xbox Game Bar recording clips in the background.'],
      windowedBoost: ['Windowed game boost', 'Flip model for borderless and windowed games: lower input delay.'],
      mouseAccel: ['Mouse acceleration off', '1:1 aim. Turns off Enhance pointer precision.'],
      stickyKeys: ['Sticky Keys popup off', 'No more Sticky Keys window when you press Shift five times.'],
      notifications: ['Pop-up notifications off', 'No app notifications popping up over your game.'],
      backgroundApps: ['Background apps off', 'Store apps stop running in the background.'],
      transparency: ['Transparency off', 'Less GPU work on the taskbar and Windows menus.'],
      animations: ['Window animations off', 'Windows open and minimize instantly.'],
    } as unknown as Dict,
  },
  theme: {
    title: 'Theme', subtitle: 'Make Reka Shade yours.', accent: 'Accent color', backdrop: 'Background', glass: 'Glass panels',
    glassHint: 'Frosted blur behind the cards.', motion: 'Animations', motionHint: 'Smooth transitions inside the app.', scale: 'Interface size',
    accents: { violet: 'Violet', magenta: 'Magenta', cyan: 'Cyan', emerald: 'Emerald', amber: 'Amber', crimson: 'Crimson', ice: 'Ice' },
    backdrops: { aurora: 'Aurora', nebula: 'Nebula', midnight: 'Midnight', oled: 'OLED black' },
  },
  com: {
    title: 'Community', subtitle: 'Swap crosshairs, reticles and presets with simple codes — no account needed.', paste: 'Paste any Reka code',
    placeholder: 'RKX1-…  ·  RKP1-…  ·  RKR1-…', import: 'Import', invalid: 'Unknown code. Reka codes start with RKX1, RKP1 or RKR1.',
    importedCrosshair: 'Crosshair imported and applied.', importedPreset: 'Preset “{name}” added to My presets.', importedReticle: 'Reticle imported.',
    picks: 'Community picks', crosshairs: 'Crosshairs', presets: 'Presets', use: 'Use', share: 'Share your setup',
    shareHint: 'Copy your current crosshair and look as codes to send to friends.', shareCrosshair: 'Copy crosshair code', shareLook: 'Copy look code',
    links: 'Get involved', github: 'Source code & releases', issues: 'Report a bug or suggest an idea',
  },
  acc: {
    title: 'Account', subtitle: 'Your local profile. Reka Shade works offline — no sign-up, no tracking.', name: 'Display name', avatar: 'Avatar',
    since: 'Using Reka Shade since {date}', stats: 'At a glance', statPresets: 'Custom presets', statCrosshairs: 'Saved crosshairs',
    statBinds: 'Active binds', statRules: 'Game profiles', backup: 'Backup & restore', backupHint: 'Move your whole setup to another PC in one file.',
    export: 'Export settings', import: 'Import settings', exported: 'Settings saved to {path}', imported: 'Settings imported.',
    privacy: 'Privacy', privacyText: 'Everything stays on this PC (%APPDATA%\\RekaShade). The only network request Reka Shade makes is the update check to GitHub.',
    reset: 'Reset everything', resetHint: 'Back to factory settings. Your Windows tweaks stay as they are.', resetConfirm: 'Reset all Reka Shade settings?', resetDone: 'All settings reset.',
  },
  set: {
    title: 'Settings', subtitle: 'How Reka Shade starts, behaves and updates.', general: 'General', language: 'Language', langAuto: 'Same as Windows',
    startWithWindows: 'Start with Windows', startMinimized: 'Start minimized in the tray', closeToTray: 'Close button keeps Reka Shade running in the tray',
    display: 'Display', applyOnStartup: 'Re-apply tuning when Reka Shade starts', restoreOnExit: 'Restore original colours when quitting',
    gpu: 'Use GPU driver vibrance (NVIDIA / AMD)', gpuDetected: 'Detected: {list}', gpuNone: 'No compatible GPU driver found — the Windows colour engine is used.',
    gpuFailed: 'Turned off after a driver problem last time. Switch it off and on to retry.', extGamma: 'Extended gamma range',
    extGammaHint: 'Lets strong looks past Windows’ gamma limit. Asks for admin rights once; sign out or restart to apply.', extGammaOn: 'Enabled', enable: 'Enable',
    updates: 'Updates', autoUpdate: 'Check for updates automatically', autoInstall: 'Install updates automatically when Reka Shade starts', checkNow: 'Check now',
    data: 'Data & support', openFolder: 'Open settings folder', diagnostics: 'Copy diagnostics', copied: 'Diagnostics copied to clipboard.',
    panic: 'Turn everything off now', panicHint: 'Disables tuning, crosshair and stretched resolution at once.',
  },
  about: {
    title: 'About', tagline: 'Display tuning, crosshair overlay and Windows optimizer — built to never touch your games.', version: 'Version {v}',
    upToDate: 'You’re up to date.', available: 'Version {v} is available.', install: 'Install & restart', downloading: 'Downloading… {pct}%',
    check: 'Check for updates', checking: 'Checking…', checkFailed: 'Couldn’t reach GitHub: {error}', notes: 'What’s new', lastChecked: 'Last checked {time}',
    safeTitle: 'Why it’s anti-cheat friendly', never: 'Never: DLL injection, reading game memory, hooks into games, simulated input or game file edits.',
    safe1: 'Tone and vibrance use your display’s own gamma table and the GPU driver — the same switches as the NVIDIA / AMD control panels.',
    safe2: 'Saturation and hue use the colour engine built into Windows (Accessibility › Colour filters).',
    safe3: 'The crosshair is an ordinary click-through window above everything, like the Discord or Xbox Game Bar overlays.',
    safe4: 'Stretched resolution changes the Windows desktop mode, exactly like Settings › Display.',
    safe5: 'Game detection reads the process list Task Manager shows. No game process is ever opened.',
    limitsTitle: 'Good to know', limits: 'Exclusive-fullscreen games draw over every overlay and over Windows’ colour engine: use borderless or windowed mode for the crosshair, hue and saturation. Tone and GPU vibrance work in every mode.',
    credits: 'Built with', updatedTo: 'Updated to Reka Shade {v}.',
  },
  legal: {
    title: 'Legal', disclaimerTitle: 'Disclaimer',
    disclaimer: 'Reka Shade is provided “as is”, without warranty of any kind. It is designed to never interact with game processes, but every game, platform and tournament sets its own rules — some competitive platforms don’t allow crosshair overlays. Check the rules of what you play.',
    trademarksTitle: 'Trademarks', trademarks: 'NVIDIA, AMD, Intel, Windows, Xbox, OBS, Discord and game names are trademarks of their owners. Reka Shade is not affiliated with or endorsed by any of them.',
    privacyTitle: 'Privacy', privacy: 'Reka Shade collects nothing. Settings stay on your PC. To check for updates it asks api.github.com for the latest release; no personal data is sent.',
    thirdParty: 'Third-party software', thirdPartyHint: 'Reka Shade is built on open-source work. Thank you to their authors.',
  },
  toast: {
    resReverted: 'Resolution reverted to native.', resTestFirst: 'Test and keep a stretched resolution in Reka Shade first.',
    resFailed: 'Couldn’t change the resolution: {error}', updateAvailable: 'Reka Shade {version} is available — see About.',
    updateFailed: 'Update failed: {error}', copied: 'Copied to clipboard.', saved: 'Saved.', panic: 'Everything is off.',
  },
}

const fr: Dict = {
  common: {
    on: 'ON', off: 'OFF', save: 'Enregistrer', cancel: 'Annuler', apply: 'Appliquer', delete: 'Supprimer', duplicate: 'Dupliquer',
    rename: 'Renommer', copy: 'Copier', copied: 'Copié !', paste: 'Coller', import: 'Importer', export: 'Exporter', reset: 'Réinitialiser',
    close: 'Fermer', add: 'Ajouter', remove: 'Retirer', custom: 'Personnalisé', none: 'Aucun', loading: 'Chargement…', error: 'Erreur',
    display: 'Écran {n}', allDisplays: 'Tous les écrans', active: 'Actif', idle: 'En veille', open: 'Ouvrir', confirm: 'Confirmer',
    minimize: 'Réduire', closeWindow: 'Fermer', enabled: 'Activé', default: 'Par défaut',
  },
  nav: {
    menu: 'MENU', general: 'GÉNÉRAL', display: 'Affichage', presets: 'Préréglages', resolution: 'Résolution', crosshair: 'Viseur',
    reticle: 'Réticule', binds: 'Raccourcis', optimizer: 'Optimiseur', theme: 'Thème', community: 'Communauté', account: 'Compte',
    settings: 'Paramètres', about: 'À propos', legal: 'Mentions légales', tuning: 'RÉGLAGE', customLook: 'Rendu perso', tuningOff: 'Réglage coupé',
    statusActive: 'Actif', statusIdle: 'En veille', update: 'Mise à jour {v}',
  },
  display: {
    title: 'Réglage de l’affichage', onDisplay: '{preset} sur {display}', off: 'Réglage désactivé — vos écrans sont exactement comme avant.',
    adjust: 'Ajuster', tone: 'Tons', color: 'Couleur', dblclick: 'Double-cliquez sur un curseur pour le réinitialiser.',
    hue: 'Teinte', saturation: 'Saturation', vibrance: 'Vibrance', red: 'Rouge', green: 'Vert', blue: 'Bleu',
    brightness: 'Luminosité', contrast: 'Contraste', gamma: 'Gamma', temperature: 'Température',
    allScreens: 'Tous les écrans', gpu: 'GPU', windowsEngine: 'Windows', startWithWindows: 'Lancer avec Windows',
    noteSat: 'La teinte et la saturation s’appliquent à tous les écrans.', resetNeutral: 'Revenir au neutre',
    preview: 'Aperçu', before: 'AVANT', after: 'APRÈS', demo: 'Démo', curve: 'Courbe',
    dragHint: 'Glissez pour comparer avant et après.', curveHint: 'Sortie des canaux rouge, vert et bleu pour chaque niveau d’entrée.',
    colorPresets: 'Préréglages de couleur', allPresets: 'Tous les préréglages', canyon: 'Canyon', neon: 'Ville néon', forest: 'Forêt',
    limited: 'Windows a limité ce rendu à {pct} % sur {display}.', unlock: 'Débloquer la plage complète',
    gammaError: '{display} refuse les changements de gamma (HDR activé ?). Les curseurs de tons n’y ont pas d’effet.',
    matrixError: 'Moteur couleur de Windows indisponible : {error}',
    vibGpu: 'La vibrance passe par votre pilote {gpu} : zéro latence, même en plein écran exclusif.',
    vibMatrix: 'Ce GPU n’a pas de vibrance pilote : Reka Shade utilise le moteur couleur de Windows.',
    rule: 'Profil de jeu actif : {exe}', welcomeTitle: 'Bienvenue dans Reka Shade',
    welcomeText: 'Bienvenue ! Choisissez un préréglage ou bougez les curseurs, puis activez l’interrupteur. Rien ici ne touche à vos jeux.',
  },
  presets: {
    title: 'Préréglages', subtitle: 'Des rendus en un clic — à partager avec un code.', saveCurrent: 'Enregistrer le rendu actuel',
    importCode: 'Importer un code', builtin: 'Intégrés', mine: 'Mes préréglages', applyTo: 'Appliquer à {display}', active: 'Actif',
    copyCode: 'Copier le code', duplicate: 'Dupliquer', rename: 'Renommer', delete: 'Supprimer', pin: 'Épingler sur Affichage',
    unpin: 'Désépingler', namePrompt: 'Nom du préréglage', newName: 'Mon préréglage {n}', saved: 'Préréglage « {name} » enregistré.',
    rules: 'Profils de jeu', rulesHint: 'Change de préréglage automatiquement quand un jeu est au premier plan. Reka Shade lit seulement la liste des fenêtres — il n’ouvre jamais le jeu.',
    addRule: 'Ajouter un jeu', pickRunning: 'Applis ouvertes', typeExe: 'ou tapez le nom d’un .exe', noRules: 'Aucun profil de jeu.',
    rulesNeedTuning: 'Les profils de jeu fonctionnent quand le réglage de l’affichage est activé.', emptyMine: 'Enregistrez votre rendu actuel pour créer votre premier préréglage.',
    pasteCode: 'Collez un code de préréglage (RKP1-…)', invalidCode: 'Ce n’est pas un code de préréglage valide.', imported: 'Préréglage « {name} » importé.',
    deleteConfirm: 'Supprimer « {name} » ?',
    tags: { neutral: 'Neutre', vivid: 'Vif', warm: 'Chaud', cool: 'Froid', cinematic: 'Cinéma', competitive: 'Compétitif', dark: 'Scènes sombres', mono: 'Mono', custom: 'Perso', shared: 'Partagé' },
  },
  res: {
    title: 'Résolution étirée', subtitle: 'Jouez en 4:3 ou à n’importe quelle taille, étirée sur tout l’écran. Seul le mode du bureau Windows change — le jeu n’est jamais touché.',
    native: 'Native {w}×{h}', stretched: 'Étirée {w}×{h}', choose: 'Choisir une résolution', display: 'Écran', custom: 'Personnalisée',
    width: 'Largeur', height: 'Hauteur', refresh: 'Fréquence', auto: 'La meilleure dispo', scaling: 'Mise à l’échelle', stretch: 'Étirée',
    aspect: 'Bandes noires', center: 'Centrée', notListed: 'Absente de la liste Windows pour cet écran. Créez-la d’abord comme résolution personnalisée dans le panneau NVIDIA / AMD / Intel.',
    test: 'Tester cette résolution', revert: 'Revenir en natif', keepTitle: 'Garder cette résolution ?', keepBody: 'Retour automatique dans {s} s sans confirmation.',
    keep: 'La garder', revertNow: 'Annuler', preview: 'Rendu sur votre écran', panel: 'Votre écran {w}×{h}', game: 'Jeu en {w}×{h}',
    autoTitle: 'Étirement automatique', autoHint: 'Passe en résolution étirée tant qu’un de ces jeux tourne, puis revient en natif à sa fermeture.',
    revertOnExit: 'Rétablir la résolution native à la fermeture de Reka Shade', needConfirm: 'Testez et gardez la résolution une fois pour débloquer l’étirement auto et le raccourci.',
    hotkey: 'Basculez à tout moment avec {key}', popular: 'Populaires', addGame: 'Ajouter un jeu', noGames: 'Aucun jeu.', confirmed: 'Prête',
    current: 'Actuelle : {w}×{h} à {hz} Hz', inGame: 'Astuce : réglez la même résolution dans le jeu, en plein écran ou sans bordure.',
  },
  xh: {
    title: 'Viseur', subtitle: 'Reste au-dessus de vos jeux et applis.', designs: 'Modèles', color: 'COULEUR', outlineColor: 'CONTOUR',
    my: 'MES VISEURS', saveCurrent: 'Enregistrer', copyCode: 'Copier le code', pasteCode: 'Coller un code',
    codesHint: 'Les codes commencent par RKX1. Copiez-en un, envoyez-le à un ami, il n’a plus qu’à cliquer sur Coller un code.', customize: 'Personnaliser',
    actualSize: 'TAILLE RÉELLE', ring: 'Anneau', arm: 'Longueur des branches', gap: 'Écart', thickness: 'Épaisseur', dot: 'Point', outline: 'Contour',
    opacity: 'Opacité', rotate: 'Pivoter de 45°', tStyle: 'Forme en T', slotHint: 'Clic pour charger · clic droit pour vider', empty: 'Vide',
    pastePrompt: 'Collez un code de viseur (RKX1-…)', invalid: 'Ce n’est pas un code de viseur valide.', savedSlot: 'Enregistré dans l’emplacement {n}.',
    slotsFull: 'Tous les emplacements sont pris — videz-en un avec un clic droit.', loaded: 'Viseur chargé.', pickColor: 'Choisir une couleur',
    reticleActive: 'Le réticule image est la superposition active. Activez ceci pour revenir au viseur.',
  },
  ret: {
    title: 'Réticule', subtitle: 'Réticules en image, et où et quand la superposition s’affiche.', library: 'Bibliothèque', upload: 'Importer une image',
    uploadHint: 'PNG ou SVG avec transparence. Redimensionné à 256 px max.', size: 'Taille', color: 'Couleur', rotation: 'Rotation', glow: 'Lueur',
    opacity: 'Opacité', yours: 'Votre image', placement: 'Placement', screen: 'Écran', offset: 'Décalage fin', offsetX: 'Horizontal',
    offsetY: 'Vertical', center: 'Recentrer', follow: 'Suivre la fenêtre active', followHint: 'Se centre sur les jeux fenêtrés plutôt que sur tout l’écran.',
    onlyFs: 'Uniquement sur les applis plein écran', onlyFsHint: 'Masqué sur le bureau et dans les fenêtres normales.', capture: 'Masquer des captures et du stream',
    captureHint: 'OBS, Discord et les enregistrements ne le verront pas (Windows 10 2004+).', hold: 'Masquer en maintenant', holdNone: 'Aucune touche — à définir dans Raccourcis',
    reasonOff: 'Superposition désactivée', reasonHold: 'Masqué tant que la touche est maintenue', reasonFs: 'En attente d’une appli plein écran', reasonImg: 'Rendu…',
    visible: 'Affiché à l’écran', behavior: 'Comportement',
    names: { holo: 'Holo', chevron: 'Chevron', 'mil-dot': 'Mil-dot', triangle: 'Triangle', bracket: 'Crochets', diamond: 'Losange', 'tactical-x': 'X tactique', 'circle-split': 'Anneau coupé' },
  },
  binds: {
    title: 'Raccourcis', subtitle: 'Des raccourcis globaux qui marchent en jeu. Les touches sont seulement lues, jamais bloquées — votre jeu les reçoit toujours.',
    toggleTuning: 'Activer/couper le réglage', toggleCrosshair: 'Afficher/masquer le viseur', nextPreset: 'Préréglage suivant', prevPreset: 'Préréglage précédent',
    toggleStretch: 'Activer/couper la résolution étirée', showApp: 'Ouvrir Reka Shade', panic: 'Panique — tout désactiver', holdHide: 'Masquer le viseur en maintenant',
    press: 'Appuyez sur une touche…', unbound: 'Non défini', clear: 'Effacer', change: 'Modifier', reset: 'Rétablir par défaut',
    holdHint: 'Astuce : associez Souris 2 (clic droit) pour masquer le viseur quand vous visez.', conflict: 'Déjà utilisé par « {action} ».',
    mouseHint: 'Les boutons de souris 2 à 5 fonctionnent aussi. Échap annule.', how: 'Comment ça marche',
    howText: 'Reka Shade lit l’état du clavier environ 60 fois par seconde, comme n’importe quelle appli en arrière-plan. Aucun hook clavier, aucune saisie simulée, rien dans le processus du jeu.',
    hold: 'MAINTIEN', toggle: 'APPUI', tip: 'Astuce',
  },
  opt: {
    title: 'Optimiseur', subtitle: 'De vrais réglages Windows pour des jeux plus fluides. Annuler remet tout comme avant.', undoAll: 'Tout annuler', optimizeAll: 'Tout optimiser',
    clean: 'Nettoyer les fichiers temporaires', cleanDesc: 'Supprime les fichiers de plus d’un jour dans votre dossier Temp. Les fichiers utilisés sont ignorés.', cleanBtn: 'Nettoyer',
    cleaned: '{files} fichiers supprimés ({size}). {skipped} fichiers utilisés ignorés.', alreadyOn: 'Déjà actif', alreadyOnHint: 'Windows l’avait déjà activé — rien à annuler.',
    win11: 'Windows 11 uniquement', failed: 'Impossible de modifier « {name} » : {error}', allDone: 'Tous les réglages sont appliqués.', allUndone: 'Tout est revenu comme avant.',
    t: {
      powerPlan: ['Mode d’alimentation hautes performances', 'Le processeur reste à pleine vitesse au lieu d’économiser. Temps d’image plus stables.'],
      gameMode: ['Mode Jeu', 'Windows donne la priorité CPU et GPU au jeu en cours.'],
      gameDVR: ['Enregistrement en arrière-plan coupé', 'La Xbox Game Bar n’enregistre plus de clips en arrière-plan.'],
      windowedBoost: ['Boost des jeux fenêtrés', 'Modèle flip pour les jeux fenêtrés et sans bordure : moins de latence.'],
      mouseAccel: ['Accélération de la souris coupée', 'Visée 1:1. Désactive « Améliorer la précision du pointeur ».'],
      stickyKeys: ['Fenêtre des touches rémanentes coupée', 'Plus de fenêtre quand vous appuyez cinq fois sur Maj.'],
      notifications: ['Notifications coupées', 'Plus de notifications d’applis par-dessus votre jeu.'],
      backgroundApps: ['Applis en arrière-plan coupées', 'Les applis du Store ne tournent plus en arrière-plan.'],
      transparency: ['Transparence coupée', 'Moins de travail GPU pour la barre des tâches et les menus.'],
      animations: ['Animations des fenêtres coupées', 'Les fenêtres s’ouvrent et se réduisent instantanément.'],
    } as unknown as Dict,
  },
  theme: {
    title: 'Thème', subtitle: 'Faites de Reka Shade le vôtre.', accent: 'Couleur d’accent', backdrop: 'Arrière-plan', glass: 'Panneaux en verre',
    glassHint: 'Flou dépoli derrière les cartes.', motion: 'Animations', motionHint: 'Transitions fluides dans l’appli.', scale: 'Taille de l’interface',
    accents: { violet: 'Violet', magenta: 'Magenta', cyan: 'Cyan', emerald: 'Émeraude', amber: 'Ambre', crimson: 'Carmin', ice: 'Glace' },
    backdrops: { aurora: 'Aurore', nebula: 'Nébuleuse', midnight: 'Minuit', oled: 'Noir OLED' },
  },
  com: {
    title: 'Communauté', subtitle: 'Échangez viseurs, réticules et préréglages avec de simples codes — sans compte.', paste: 'Collez un code Reka',
    placeholder: 'RKX1-…  ·  RKP1-…  ·  RKR1-…', import: 'Importer', invalid: 'Code inconnu. Les codes Reka commencent par RKX1, RKP1 ou RKR1.',
    importedCrosshair: 'Viseur importé et appliqué.', importedPreset: 'Préréglage « {name} » ajouté à Mes préréglages.', importedReticle: 'Réticule importé.',
    picks: 'Sélection de la communauté', crosshairs: 'Viseurs', presets: 'Préréglages', use: 'Utiliser', share: 'Partager votre config',
    shareHint: 'Copiez votre viseur et votre rendu actuels en codes à envoyer à vos amis.', shareCrosshair: 'Copier le code du viseur', shareLook: 'Copier le code du rendu',
    links: 'Participer', github: 'Code source et versions', issues: 'Signaler un bug ou proposer une idée',
  },
  acc: {
    title: 'Compte', subtitle: 'Votre profil local. Reka Shade marche hors ligne — sans inscription, sans pistage.', name: 'Pseudo', avatar: 'Avatar',
    since: 'Utilise Reka Shade depuis le {date}', stats: 'En bref', statPresets: 'Préréglages perso', statCrosshairs: 'Viseurs enregistrés',
    statBinds: 'Raccourcis actifs', statRules: 'Profils de jeu', backup: 'Sauvegarde et restauration', backupHint: 'Transférez toute votre config sur un autre PC en un fichier.',
    export: 'Exporter les paramètres', import: 'Importer des paramètres', exported: 'Paramètres enregistrés dans {path}', imported: 'Paramètres importés.',
    privacy: 'Confidentialité', privacyText: 'Tout reste sur ce PC (%APPDATA%\\RekaShade). La seule requête réseau de Reka Shade est la recherche de mises à jour sur GitHub.',
    reset: 'Tout réinitialiser', resetHint: 'Retour aux réglages d’usine. Vos réglages Windows restent tels quels.', resetConfirm: 'Réinitialiser tous les paramètres de Reka Shade ?', resetDone: 'Tous les paramètres sont réinitialisés.',
  },
  set: {
    title: 'Paramètres', subtitle: 'Comment Reka Shade démarre, se comporte et se met à jour.', general: 'Général', language: 'Langue', langAuto: 'Comme Windows',
    startWithWindows: 'Lancer avec Windows', startMinimized: 'Démarrer réduit dans la zone de notification', closeToTray: 'Le bouton fermer garde Reka Shade dans la zone de notification',
    display: 'Affichage', applyOnStartup: 'Réappliquer le réglage au lancement de Reka Shade', restoreOnExit: 'Restaurer les couleurs d’origine en quittant',
    gpu: 'Utiliser la vibrance du pilote GPU (NVIDIA / AMD)', gpuDetected: 'Détecté : {list}', gpuNone: 'Aucun pilote GPU compatible — le moteur couleur de Windows est utilisé.',
    gpuFailed: 'Coupé après un problème de pilote la dernière fois. Désactivez puis réactivez pour réessayer.', extGamma: 'Plage gamma étendue',
    extGammaHint: 'Laisse passer les rendus forts au-delà de la limite gamma de Windows. Demande les droits admin une fois ; déconnectez-vous ou redémarrez pour appliquer.', extGammaOn: 'Activée', enable: 'Activer',
    updates: 'Mises à jour', autoUpdate: 'Rechercher les mises à jour automatiquement', autoInstall: 'Installer les mises à jour automatiquement au lancement', checkNow: 'Vérifier maintenant',
    data: 'Données et support', openFolder: 'Ouvrir le dossier des paramètres', diagnostics: 'Copier le diagnostic', copied: 'Diagnostic copié dans le presse-papiers.',
    panic: 'Tout désactiver maintenant', panicHint: 'Coupe le réglage, le viseur et la résolution étirée d’un coup.',
  },
  about: {
    title: 'À propos', tagline: 'Réglage de l’affichage, viseur en superposition et optimiseur Windows — conçu pour ne jamais toucher à vos jeux.', version: 'Version {v}',
    upToDate: 'Vous êtes à jour.', available: 'La version {v} est disponible.', install: 'Installer et redémarrer', downloading: 'Téléchargement… {pct} %',
    check: 'Rechercher des mises à jour', checking: 'Vérification…', checkFailed: 'Impossible de joindre GitHub : {error}', notes: 'Nouveautés', lastChecked: 'Dernière vérification {time}',
    safeTitle: 'Pourquoi c’est compatible anti-triche', never: 'Jamais : injection de DLL, lecture de la mémoire du jeu, hooks dans le jeu, saisie simulée ni modification des fichiers du jeu.',
    safe1: 'Les tons et la vibrance passent par la table gamma de l’écran et le pilote GPU — les mêmes réglages que les panneaux NVIDIA / AMD.',
    safe2: 'La saturation et la teinte utilisent le moteur couleur intégré à Windows (Accessibilité › Filtres de couleur).',
    safe3: 'Le viseur est une simple fenêtre transparente aux clics, au-dessus de tout, comme les superpositions Discord ou Xbox Game Bar.',
    safe4: 'La résolution étirée change le mode du bureau Windows, exactement comme Paramètres › Affichage.',
    safe5: 'La détection des jeux lit la liste des processus affichée par le Gestionnaire des tâches. Aucun processus de jeu n’est jamais ouvert.',
    limitsTitle: 'Bon à savoir', limits: 'Les jeux en plein écran exclusif passent par-dessus toute superposition et le moteur couleur de Windows : utilisez le mode sans bordure ou fenêtré pour le viseur, la teinte et la saturation. Les tons et la vibrance GPU marchent dans tous les modes.',
    credits: 'Conçu avec', updatedTo: 'Mis à jour vers Reka Shade {v}.',
  },
  legal: {
    title: 'Mentions légales', disclaimerTitle: 'Avertissement',
    disclaimer: 'Reka Shade est fourni « tel quel », sans aucune garantie. Il est conçu pour ne jamais interagir avec les processus de jeu, mais chaque jeu, plateforme et tournoi a ses propres règles — certaines plateformes compétitives interdisent les viseurs en superposition. Vérifiez les règles de ce à quoi vous jouez.',
    trademarksTitle: 'Marques', trademarks: 'NVIDIA, AMD, Intel, Windows, Xbox, OBS, Discord et les noms de jeux sont des marques de leurs propriétaires. Reka Shade n’est affilié à aucun d’eux ni approuvé par eux.',
    privacyTitle: 'Confidentialité', privacy: 'Reka Shade ne collecte rien. Les paramètres restent sur votre PC. Pour les mises à jour, il demande la dernière version à api.github.com ; aucune donnée personnelle n’est envoyée.',
    thirdParty: 'Logiciels tiers', thirdPartyHint: 'Reka Shade repose sur des projets open source. Merci à leurs auteurs.',
  },
  toast: {
    resReverted: 'Résolution native rétablie.', resTestFirst: 'Testez et gardez d’abord une résolution étirée dans Reka Shade.',
    resFailed: 'Impossible de changer la résolution : {error}', updateAvailable: 'Reka Shade {version} est disponible — voir À propos.',
    updateFailed: 'Échec de la mise à jour : {error}', copied: 'Copié dans le presse-papiers.', saved: 'Enregistré.', panic: 'Tout est désactivé.',
  },
}

const dicts: Record<'en' | 'fr', Dict> = { en, fr }

export const lang = ref<'en' | 'fr'>('en')
export const locale = computed(() => (lang.value === 'fr' ? 'fr-FR' : 'en-US'))

function lookup(d: Dict, key: string): unknown {
  let cur: unknown = d
  for (const part of key.split('.')) {
    if (cur && typeof cur === 'object' && part in (cur as Dict)) cur = (cur as Dict)[part]
    else return undefined
  }
  return cur
}

/** Translate a key; {name} placeholders are filled from params. */
export function t(key: string, params?: Record<string, unknown>): string {
  let v = lookup(dicts[lang.value], key)
  if (typeof v !== 'string') v = lookup(en, key)
  if (typeof v !== 'string') return key
  if (!params) return v
  return v.replace(/\{(\w+)\}/g, (_, k) => (params[k] !== undefined ? String(params[k]) : `{${k}}`))
}

/** Optimizer tweak name/description pair. */
export function tweakText(id: string): [string, string] {
  const v = (lookup(dicts[lang.value], `opt.t.${id}`) ?? lookup(en, `opt.t.${id}`)) as unknown
  return Array.isArray(v) ? [String(v[0]), String(v[1])] : [id, '']
}

export function fmtBytes(n: number): string {
  const u = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (n >= 1024 && i < u.length - 1) { n /= 1024; i++ }
  return `${n.toFixed(i ? 1 : 0)} ${u[i]}`
}
