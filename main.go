package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

type BadgerModelData struct {
	FormatVersion     string `json:"format_version"`
	MinecraftGeometry []struct {
		Bones []struct {
			//ignore the weird unicode escaped nonsense for now.
			Name   string `json:"name"`
			Parent string `json:"parent"`
			Pivot  []int  `json:"pivot"`
			Scale  []int  `json:"scale"`
		} `json:"bones"`
		Meshes []struct {
			MetaMaterial string        `json:"meta_material"`
			NormalSets   [][][]float64 `json:"normal_sets"`
			Positions    [][]float64   `json:"positions"`
			Triangles    []int         `json:"triangles"`
			UvSets       [][][]float64 `json:"uv_sets"`
			Weights      [][]int       `json:"weights"`
		} `json:"meshes"`
		Description struct {
			Identifier string `json:"identifier"`
		} `json:"description"`
	} `json:"minecraft:geometry"`
}

func (b *BadgerModelData) toObj() string {
	objfile := "g mcl_model\n"
	meshIndexOffset := 0
	for _, bMesh := range b.MinecraftGeometry[0].Meshes {

		for _, vertex := range bMesh.Positions {
			objfile += fmt.Sprintf("v %v %v %v\n", vertex[0], vertex[1], vertex[2])
		}
		for _, uv := range bMesh.UvSets[0] {
			objfile += fmt.Sprintf("vt %v %v\n", uv[0], 1-uv[1])
		}
		for _, normal := range bMesh.NormalSets[0] {
			objfile += fmt.Sprintf("vn %v %v %v\n", normal[0], normal[1], normal[2])
		}
		for i := 0; i < len(bMesh.Triangles); i += 3 {
			p1 := bMesh.Triangles[i] + 1 + meshIndexOffset
			p2 := bMesh.Triangles[i+1] + 1 + meshIndexOffset
			p3 := bMesh.Triangles[i+2] + 1 + meshIndexOffset
			//forgive me
			objfile += fmt.Sprintf("f %v/%v/%v %v/%v/%v %v/%v/%v\n", p1, p1, p1, p2, p2, p2, p3, p3, p3)
		}
		meshIndexOffset += len(bMesh.Positions)
	}
	return objfile
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "obj",
				Aliases: []string{"o"},
				Usage:   "bmesh obj path/to/model.json path/to/output.obj",
				Action: func(c *cli.Context) error {
					file, err := os.ReadFile(c.Args().First())
					if err != nil {
						panic(err)
					}
					var model BadgerModelData
					json.Unmarshal(file, &model)
					os.WriteFile(c.Args().Get(1), []byte(model.toObj()), 0667)
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
