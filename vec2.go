package main

import "math"

type vec2 struct {
	x float64
	y float64
}

func (v *vec2) LengthSquared() float64 {
	return v.x*v.x + v.y*v.y
}

func (v *vec2) Add(u *vec2) vec2 {
	return vec2{
		v.x + u.x,
		v.y + u.y}
}

func (v *vec2) Sub(u *vec2) vec2 {
	return vec2{
		v.x - u.x,
		v.y - u.y}
}

func (v *vec2) Dot(u *vec2) float64 {
	return v.x*u.x + v.y*u.y
}

func (v *vec2) Normalize() vec2 {
	length := math.Sqrt(v.LengthSquared())
	return vec2{
		v.x / length,
		v.y / length}
}
