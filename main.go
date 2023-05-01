package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

type BadgerMesh struct {
	MetaMaterial string        `json:"meta_material"`
	NormalSets   [][][]float64 `json:"normal_sets"`
	Positions    [][]float64   `json:"positions"`
	Triangles    []int         `json:"triangles"`
	UvSets       [][][]float64 `json:"uv_sets"`
	Weights      [][]int       `json:"weights"`
}

type BadgerGeometry struct {
	Bones []struct {
		//ignore the weird unicode escaped nonsense for now.
		Name   string `json:"name"`
		Parent string `json:"parent"`
		Pivot  []int  `json:"pivot"`
		Scale  []int  `json:"scale"`
	} `json:"bones"`
	Meshes      []BadgerMesh `json:"meshes"`
	Description struct {
		Identifier string `json:"identifier"`
	} `json:"description"`
}

type BadgerModelData struct {
	FormatVersion     string           `json:"format_version"`
	MinecraftGeometry []BadgerGeometry `json:"minecraft:geometry"`
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

func _flt(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

func _int(str string) int {
	f, _ := strconv.Atoi(str)
	return f
}

func (b *BadgerModelData) fromObj(obj string) {
	obj = strings.Replace(obj, "\r", "", -1)
	b.FormatVersion = "1.14.0"

	geo := BadgerGeometry{}
	geo.Description.Identifier = "geometry.bmesh_obj_import"

	mesh := BadgerMesh{
		MetaMaterial: "mat_debug_shadercubes", //placeholder
	}

	var uvs [][]float64
	var normals [][]float64
	var positions [][]float64
	for _, line := range strings.Split(obj, "\n") {
		parts := strings.Split(line, " ")
		if strings.HasPrefix(line, "v ") {
			positions = append(positions,
				[]float64{
					_flt(parts[1]),
					_flt(parts[2]),
					_flt(parts[3]),
				},
			)
		}
		if strings.HasPrefix(line, "vt ") {
			uvs = append(uvs, []float64{
				_flt(parts[1]),
				1 - _flt(parts[2]),
			})
		}
		if strings.HasPrefix(line, "vn ") {
			normals = append(normals,
				[]float64{
					_flt(parts[1]),
					_flt(parts[2]),
					_flt(parts[3]),
					1,
				},
			)
		}
	}

	var mUVs [][]float64
	var mNormals [][]float64
	var mPositions [][]float64

	currentTriangleIndex := 0

	for _, line := range strings.Split(obj, "\n") {
		if strings.HasPrefix(line, "f ") {
			parts := strings.Split(line, " ")
			for j := 1; j < 4; j++ {
				subpieces := strings.Split(parts[j], "/")
				pos := positions[_int(subpieces[0])-1]
				mPositions = append(mPositions, pos)
				mNormals = append(mNormals, normals[_int(subpieces[2])-1])
				mUVs = append(mUVs, uvs[_int(subpieces[1])-1])
				//place the correct indices into all arrays.
				mesh.Triangles = append(mesh.Triangles, currentTriangleIndex)
				currentTriangleIndex++
			}
		}
	}

	//	fmt.Printf("%v %v\n", uvs, normals)
	for i := 0; i < len(mPositions); i++ {
		mesh.Weights = append(mesh.Weights, []int{1})
	}

	mesh.UvSets = append(mesh.UvSets, mUVs)
	mesh.Positions = mPositions
	mesh.NormalSets = append(mesh.NormalSets, mNormals)
	geo.Meshes = append(geo.Meshes, mesh)
	b.MinecraftGeometry = append(b.MinecraftGeometry, geo)
	//validation
	l := len(mesh.Positions)
	fmt.Printf("u: %v n: %v w: %v p: %v\n", len(mesh.UvSets[0]), len(mesh.NormalSets[0]), len(mesh.Weights), len(mesh.Positions))
	if len(mesh.UvSets[0]) != l || len(mesh.NormalSets[0]) != l || len(mesh.Weights) != l {
		panic("OBJ incorrectly converted! Mismatched array lengths.")
	}
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
			{
				Name:  "bo",
				Usage: "<todo>",
				Action: func(c *cli.Context) error {
					file, err := os.ReadFile(c.Args().First())
					if err != nil {
						panic(err)
					}
					var model BadgerModelData
					model.fromObj(string(file))
					out, _ := json.Marshal(model)
					os.WriteFile(c.Args().Get(1), []byte(out), 0667)
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
