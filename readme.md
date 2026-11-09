# Projekt

Web-Anwendung für das Backend einer Fahrradverleih Firma.

## Anforderungen

### Anforderung 0

Es soll ein Web-Backend erstellt werden. Zentraler Punkt des Backends ist
ein Dashboard. Am oberen Rand gibt es stilisierte Tab-Reiter, die auf die
Unterseiten (wie etwa Anforderung 1, Anforderung 4) umschalten lassen.

### Anforderung 1

**User Story:** Ein Mitarbeiter kann die Fahrrad-Datenbank verwalten, indem er
neue Fahrräder hinzufügt oder bestehende Fahrräder löscht. Die Liste der im System
vorhandenen Fahrräder wird angezeigt. Ist ein Fahrrad verliehen, dann wird der Name
des Kunden dazu angezeigt.
Wird ein neues Fahrrad hinzugefügt, soll die Inventarnummer automatisch
vorgeschlagen werden (Siehe Anforderung 2 für das Format).
Für diese User Story soll eine eigene Seite erstellt werden.

### Anforderung 2

Die Fahrrad-Datenbank besteht aus Einträgen mit foldengenden Feldern:
* Inventarnummer (<jahr>-<nummer> ... wobei <jahr> das Anschaffungsjahr ist,
und <nummer> eine fortlaufende Nummer.
* Marke (Freitext)
* Typ (Textauswahl zwischen "Rennrad", "Mountain-Bike", etc. - Bitte Liste selbst
  sinnvoll ergänzen)
* Modell (Freitext)
* Sattelhöhe cm mit Minimum und Maximum, da verstellbar (Zahl)
* Raddurchmesser in Zoll (Zahl)

### Anforderung 3

**User Story:** Ein Mitarbeiter kann die Kunden-Datenbank verwalten. Wie bei
Anforderung 1 soll er erstellen können, jedoch nicht löschen. Ein Kunde kann
gesperrt und entsperrt werden. Ein gesperrter Kunde kann keine Fahrräder mehr
ausleihen.

### Anforderung 4

Die Kunden-Datenbank besteht aus Einträgen mit folgenden Feldern:
* Name
* Telefonnummer
* E-Mail
* Anschrift
* Führerschein oder Passnummer (mit selektion, ob die eingegebene Nummer ein
  Führerschein oder ein Pass ist)
* Foto-Upload für ein Foto eines Ausweises
* Bevorzugte Sattelhöhe mit min und max.

Beim Anlegen eines Kunden muss entweder die Führerschein- bzw. Passnummer
oder das Ausweisfoto angegeben werden. Eines von beiden reicht, es können
aber auch beide eingetragen werden.

### Anforderung 5

**User Story:** Ein Mitarbeiter soll einem Kunden ein Fahrrad verleihen
können. Dazu wählt er einen Kunden und bekommt die Fahrräder aufgelistet,
bei denen die Sattelhöhe auf die bevorzuge Sattelhöhe des Kunden einstellbar
ist. Es werden nur die Fahrräder aufgelistet, die nicht bereits verliehen sind. Aus
dieser Liste kann der Mitarbeiter ein Fahrrad auswählen. Dieses Fahrrad wird als
Verliehen an den Kunden registriert.

Die Granularität der Ausleihe beträgt Stunden. Das heisst, es werden immer
ganze Stunden im Sinne der Uhrzeit als mögliche Verleihblöcke angesehen.
Sowohl Ausleihe als auch Reservierung beziehen sich auf diese Stundenblöcke.

### Anforderung 6

**User Story:** Ein Mitarbeiter soll einen verliehenes Fahrrad von einem Kunden
wieder zurücknehmen können. Dazu hat er einen eigenen Schirm, in dem er
entweder einen Kunden oder ein Fahrrad auswählen kann. Nach Auswahl wird das
jeweilige passende Gegenstück (zB Kunde, der das Fahrrad ausgeliehen hat, oder
das Fahrrad des Kunden) angezeigt. Mittels Auswahl kann die Ausleihe beendet
werden.

### Anforderung 7

**User Story:** Ein Kunde soll über eine eigene Frontend-Seite ein Fahrrad für sich
reservieren können. Eine Reservierung ist ein zeitlich terminiertes Ausleihen, welches
in der Zukunft liegen muss. Bei einer Reservierung werden die Ski tagesweise an die
Kunden vergeben.

Klarstellungen:
* Der Kunde wählt sich auf der Kundenseite selbst aus einer Namensliste (keine
  Anmeldung). Die Liste enthält nur Namen, keine Kontakt- oder Ausweisdaten.
* Reservierungen sind frühestens ab morgen möglich und gelten von–bis
  einschließlich. Ein Fahrrad kann nur einmal gleichzeitig vergeben werden.
* Beim Verleih (Anforderung 5) muss der Mitarbeiter ein Rückgabedatum und
  die Rückgabezeit eintragen. Ein Fahrrad ist im Zeitraum einer Ausleihe nicht
  reservierbar, und ein für einen anderen Kunden reserviertes Fahrrad kann nicht
  über die Reservierung hinweg verliehen werden. Ein überfälliges Fahrrad ist nicht
  reservierbar, bis es zurück ist.

### Anforderung 8

Wird ein Fahrrad reserviert, ist das in der Liste der Fahrräder ersichtlich: Statt
„verfügbar“ steht „reserviert ab <datum>-<uhrzeit>“ (erstes Stunde der nächsten
Reservierung). Das Fahrrad kann weiterhin ausgeliehen werden, aber nur bis
zur Stunde vor der ersten reservierten Stunde.

Klarstellungen:
* Ist ein Fahrrad gerade verliehen und zusätzlich reserviert, werden beide
  Informationen angezeigt.
* Im Verleih (Anforderung 5) wird bei einem reservierten Fahrrad angezeigt, bis
  wann er höchstens verliehen werden kann. Ein späteres Rückgabedatum wird
  abgelehnt.
* Der Kunde, der das Fahrrad reserviert hat, kann ihn auch über seine eigene
  Reservierung hinweg ausleihen (Abholung der Reservierung).
