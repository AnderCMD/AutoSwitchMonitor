// Command autoswitchmonitor corre en segundo plano (bandeja del sistema) y:
//
//  1. Vigila un dispositivo USB específico (el que se mueve con el switch
//     KVM) y, cuando aparece, cambia automáticamente la entrada de video
//     del monitor a la que corresponde a esta PC (DDC/CI).
//  2. Registra combinaciones de teclas globales configurables para forzar
//     el cambio de entrada manualmente (útil para la PC que no está
//     conectada al KVM, o para saltar a cualquier entrada en cualquier
//     momento).
//
// Cada PC corre su propia instancia con su propio config.yaml (ver
// internal/config). Usa `-scan` para identificar el vendor_id/product_id
// del dispositivo USB que debes vigilar.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.design/x/hotkey/mainthread"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/config"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/trayapp"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/usbwatch"
)

func main() {
	scan := flag.Bool("scan", false, "Lista los dispositivos USB conectados cada segundo, para identificar el vendor_id/product_id del dispositivo que cambia con el KVM (enciende/apaga el KVM entre PCs mientras corre esto y compara la lista)")
	printConfigPath := flag.Bool("config-path", false, "Imprime la ruta del archivo config.yaml y termina")
	flag.Parse()

	if *printConfigPath {
		path, err := config.Path()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)
		return
	}

	if *scan {
		runScan()
		return
	}

	cfg, path, err := config.Load()
	if err != nil {
		log.Fatalf("cargando configuración: %v", err)
	}
	fmt.Fprintf(os.Stderr, "AutoSwitchMonitor usando config: %s\n", path)

	// golang.design/x/hotkey necesita que el registro y el loop de eventos
	// corran en el hilo principal del SO (crítico en macOS por CGEventTap).
	mainthread.Init(func() {
		trayapp.Run(cfg, path)
	})
}

func runScan() {
	fmt.Println("Escaneando dispositivos USB. Cambia el KVM entre tus PCs y observa qué")
	fmt.Println("vendor_id:product_id aparece/desaparece. Ctrl+C para salir.")
	fmt.Println()

	prev := map[string]usbwatch.Device{}
	for {
		devices, err := usbwatch.List()
		if err != nil {
			log.Fatalf("listando dispositivos USB: %v", err)
		}

		curr := map[string]usbwatch.Device{}
		for _, d := range devices {
			key := d.VendorID + ":" + d.ProductID
			curr[key] = d
		}

		for key, d := range curr {
			if _, existed := prev[key]; !existed {
				fmt.Printf("+ CONECTADO   %-4s:%-4s  %s\n", d.VendorID, d.ProductID, d.Name)
			}
		}
		for key, d := range prev {
			if _, still := curr[key]; !still {
				fmt.Printf("- DESCONECTADO %-4s:%-4s  %s\n", d.VendorID, d.ProductID, d.Name)
			}
		}

		prev = curr
		time.Sleep(1 * time.Second)
	}
}
