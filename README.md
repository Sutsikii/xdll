# xdll

Petite app Windows pour télécharger les images et vidéos d'un post X/Twitter, d'un Short (ou d'une vidéo) YouTube d'un post/reel Instagram ou d'une épingle Pinterest, et les garder dans une bibliothèque locale.

Faite avec [MyGo](https://github.com/egoist/mygo) (Go + WebView2).

## Fonctions

- Coller l'URL d'un post (champ du haut, Ctrl+V dans la fenêtre ou glisser un lien)
- X : images en taille originale, vidéos et GIF en MP4 meilleure qualité
- YouTube : meilleure qualité H.264 avec la piste audio d'origine (pas le doublage auto)
- Instagram : photos, vidéos, carrousels
- Pinterest : images, vidéos, épingles idées (liens pinterest.com/pin/… et pin.it/…)
- Le site d'origine est affiché sur chaque post (icône), filtrable dans la barre d'outils
- Dossiers et sous-dossiers (clic droit → Nouveau sous-dossier, ou glisser un dossier sur un autre), favoris, notes
- Recherche, filtre par type, tri
- Sélection multiple, export, suppression
- Clic droit sur un post, le fond, la visionneuse ou la barre latérale : les actions utiles là où on clique (copier le fichier pour le coller dans l'Explorateur ou une discussion, déplacer, favori, tri…)
- Copier pour Claude : texte du post (avec les chemins des fichiers), image, ou les deux
- Paramètres (en bas à gauche) : thème clair, sombre ou celui de Windows, place prise par site et par type, qualité des vidéos, sauvegarde et import de toute la bibliothèque

Les médias restent sur le PC, même si le post est supprimé de X.

## Données

Par défaut tout est dans `%APPDATA%\xdll` :

- `library.json` : posts, dossiers, favoris, notes
- `media\<id du post>\` : les fichiers

Paramètres → Données → **Changer de dossier** déplace `library.json` et `media` ailleurs (autre disque, dossier synchronisé…). Si le dossier choisi contient déjà une bibliothèque xdll, elle est utilisée telle quelle. Le chemin choisi est retenu dans `%APPDATA%\xdll\data.json` ; s'il est introuvable au lancement (disque débranché), l'app le signale et utilise le dossier par défaut en attendant.

## Qualité des vidéos

Paramètres → Qualité des vidéos :

- **Économie** (par défaut) : X en 720p, YouTube en AV1 jusqu'à 1080p. 2 à 4 fois moins de place, sans réencodage : ce sont des versions que les sites servent déjà. Instagram et Pinterest n'ont qu'une version.
- **Maximale** : la meilleure version du site, en H.264.

**Alléger** re-télécharge en version économie les vidéos enregistrées en qualité maximale, et garde chaque nouveau fichier s'il est plus léger. Les posts qui ne sont plus en ligne gardent leurs fichiers.

À l'export, les vidéos en version économie sont re-téléchargées en qualité maximale tant que le post est en ligne (option à décocher pour exporter les fichiers tels quels).

Les vidéos YouTube en AV1 se lisent dans l'app ; pour les ouvrir dans le lecteur Windows, il faut l'extension AV1 du Microsoft Store (souvent déjà installée sur Windows 11).

## Changer de PC

1. Paramètres → **Sauvegarder la bibliothèque** : un `.zip` avec les posts, dossiers, favoris, notes, médias et réglages.
2. Copier le `.zip` sur le nouveau PC.
3. Installer xdll, Paramètres → **Remplacer** (ou **Fusionner** pour ajouter seulement les posts absents).

## Mises à jour

L'app se met à jour toute seule depuis les [releases GitHub](https://github.com/Sutsikii/xdll/releases) : une vérification par jour, avec une fenêtre qui propose d'installer. Dans Paramètres : vérifier maintenant, désactiver la vérification, ou installer sans demander.

## Raccourcis

| Touche | Action |
|---|---|
| Ctrl+V | Ajouter le post du presse-papier |
| Ctrl+L | Aller au champ URL |
| Ctrl+F | Rechercher |
| Ctrl+A | Tout sélectionner |
| Suppr | Supprimer la sélection |
| ← → | Média précédent / suivant (visionneuse) |
| F | Favori (visionneuse) |
| C / T / I | Copier pour Claude : tout / texte / image (visionneuse) |
| Échap | Fermer / désélectionner |

## Développement

Il faut Go 1.27 ou plus.

```sh
go tool mygo dev     # lancer en dev
go tool mygo build   # build/windows-amd64/xdll.exe + installateur
go test ./...
XDLL_NET=1 go test ./...   # avec de vrais téléchargements
XDLL_CLIP=1 go test -run Clipboard .   # vrai presse-papier (Windows)
```

### Publier une version

1. Changer `version` dans `mygo.json` et ajouter sa section dans `CHANGELOG.md` (affichée dans la fenêtre de mise à jour).
2. Commit, puis `git tag v0.3.0 && git push origin main v0.3.0`.

Le workflow `Release` teste, build l'installateur signé et publie la release : les apps installées la proposent dans la journée.

La clé qui signe les mises à jour est dans `%APPDATA%\mygo\update-keys\mygo-update.key` et dans le secret `MYGO_UPDATER_PRIVATE_KEY` du dépôt. Sans elle, les apps installées ne peuvent plus être mises à jour : la garder aussi dans un gestionnaire de mots de passe.

## Structure

| Fichier | Rôle |
|---|---|
| `main.go` | fenêtre, sert la page et les médias |
| `sources.go` | reconnaît l'URL (X, YouTube, Instagram, Pinterest) et télécharge les médias |
| `twitter.go` | récupère le post X (API syndication, puis FxTwitter) et télécharge |
| `youtube.go` | récupère la vidéo YouTube (API du lecteur, client Android) |
| `instagram.go` | récupère le post Instagram (via des serveurs InstaFix) |
| `pinterest.go` | récupère l'épingle Pinterest (API du site, sans compte) |
| `mux.go` | fusionne la vidéo et le son YouTube en un MP4, sans ffmpeg |
| `library.go` | bibliothèque : posts, dossiers, export |
| `claude.go` | copie pour Claude (texte + image) |
| `optimize.go` | allège la bibliothèque (vidéos re-téléchargées en économie) |
| `backup.go` | sauvegarde et import (.zip) |
| `update.go` | réglages des mises à jour |
| `datadir.go` | choix du dossier des données |
| `theme.go` | thème clair, sombre ou du système, retenu dans `%APPDATA%\xdll\theme.json` |
| `web/` | interface (HTML, CSS, JS sans build) |

## Limites

Les posts protégés, supprimés ou marqués sensibles ne sont pas accessibles sans compte.

YouTube, Instagram et Pinterest ne sont pas prévus pour ça et changent souvent : un site peut cesser de marcher sans prévenir. Instagram passe par des serveurs tiers (InstaFix) : pas de date de publication, légende parfois coupée, photos en taille réduite. Les vidéos Pinterest sont en 720p.
