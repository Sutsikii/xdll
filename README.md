# xdll

Petite app Windows pour télécharger les images et vidéos d'un post X/Twitter et les garder dans une bibliothèque locale.

Faite avec [MyGo](https://github.com/egoist/mygo) (Go + WebView2).

## Fonctions

- Coller l'URL d'un post (champ du haut, Ctrl+V dans la fenêtre ou glisser un lien)
- Images en taille originale, vidéos et GIF en MP4 meilleure qualité
- Dossiers, favoris, notes
- Recherche, filtre par type, tri
- Sélection multiple, export, suppression
- Copier pour Claude : texte du post (avec les chemins des fichiers), image, ou les deux
- Paramètres (roue en bas à gauche) : sauvegarde et import de toute la bibliothèque

Les médias restent sur le PC, même si le post est supprimé de X.

## Données

Tout est dans `%APPDATA%\xdll` :

- `library.json` : posts, dossiers, favoris, notes
- `media\<id du post>\` : les fichiers

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
| `twitter.go` | récupère le post (API syndication, puis FxTwitter) et télécharge |
| `library.go` | bibliothèque : posts, dossiers, export |
| `claude.go` | copie pour Claude (texte + image) |
| `backup.go` | sauvegarde et import (.zip) |
| `update.go` | réglages des mises à jour |
| `web/` | interface (HTML, CSS, JS sans build) |

## Limites

Les posts protégés, supprimés ou marqués sensibles ne sont pas accessibles sans compte.
