# Rapport — Étape 8 : Zoom Out intermédiaire

*   **Relecteur** : egm-zoom-out (Skeptical Architect), 2e passe haute altitude
*   **Entrées** : rapports des étapes 5, 6 et 7 + fichier complet
*   **Date** : 2026-11-23

## Verdict global

Le design tient globalement en haute altitude, mais présente des problèmes d'ordre architectural nécessitant des corrections. Les préoccupations antérieures sur les détails de bas niveau dans la section Technical Plan et sur le risque de fuite du code de test vers la production sont confirmées comme architecturales (et non purement éditoriales).

## Constats

1. **Sévérité** : Major
   **Portée affectée** : section « Technical Plan (Architecture) » (lignes 37-55)
   **Justification** : la section contient des détails d'implémentation de bas niveau (noms de fichiers helpers comme `internal/testutil/postgres.go`, signatures telles que `NewUserFixture`, procédures d'installation spécifiques) qui détournent l'attention de l'architecture et des flux de données, contrevenant à la règle anti-weeds du standard zoom-out.

2. **Sévérité** : Major
   **Portée affectée** : mentions de `internal/testutil` (lignes 22, 41, 42, 89, 90, 93, 94, 97, 98) et déclaration de compatibilité (ligne 76)
   **Justification** : le plan suppose que `internal/testutil` reste réservé aux tests sans mécanisme de vérification des imports de production. Une fuite polluerait le binaire (taille, dépendances de test en production). Aucune sauvegarde (ex. extension du garde-fou « legacy drivers » de la CI couvrant `internal/testutil`) n'est mentionnée.

3. **Sévérité** : Nit
   **Portée affectée** : section « Alternatives Considered » (lignes 57-71)
   **Justification** : les rejets restent attribués à « Rejected by author » sans expliciter les inconvénients techniques (complexité de maintenance, surcharge, couplage), malgré quelques éléments de raisonnement présents.

4. **Sévérité** : Nit
   **Portée affectée** : lignes 92 et 114 (retrofit de `tests/article_test.go`, fixtures vs RLS)
   **Justification** : le plan mentionne la préservation de l'isolation RLS mais ne nomme pas explicitement les tests RLS fail-closed existants (`tests/article_test.go` lignes ~658-711 : insert cross-tenant 42501, 0 ligne sans contexte) comme garants à préserver — omission éditoriale qui affaiblit l'assurance architecturale.
