# xdll

Petite app Windows pour télécharger les images et vidéos d'un post X/Twitter et les garder dans une bibliothèque locale.

Faite avec [MyGo](https://github.com/egoist/mygo) (Go + WebView2).

## Fonctions

- Coller l'URL d'un post (champ du haut, Ctrl+V dans la fenêtre ou glisser un lien)
- Images en taille originale, vidéos et GIF en MP4 meilleure qualité
- Dossiers, favoris, notes
- Recherche, filtre par type, tri
- Sélection multiple, export, suppression

Les médias restent sur le PC, même si le post est supprimé de X.

## Données

Tout est dans `%APPDATA%\xdll` :

- `library.json` : posts, dossiers, favoris, notes
- `media\<id du post>\` : les fichiers

Sauvegarder ce dossier suffit pour garder la bibliothèque.

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
| Échap | Fermer / désélectionner |

## Développement

Il faut Go 1.27 ou plus.

```sh
go tool mygo dev     # lancer en dev
go tool mygo build   # build/windows-amd64/xdll.exe + installateur
go test ./...
XDLL_NET=1 go test ./...   # avec de vrais téléchargements
```

## Structure

| Fichier | Rôle |
|---|---|
| `main.go` | fenêtre, sert la page et les médias |
| `twitter.go` | récupère le post (API syndication, puis FxTwitter) et télécharge |
| `library.go` | bibliothèque : posts, dossiers, export |
| `web/` | interface (HTML, CSS, JS sans build) |

## Limites

Les posts protégés, supprimés ou marqués sensibles ne sont pas accessibles sans compte.
