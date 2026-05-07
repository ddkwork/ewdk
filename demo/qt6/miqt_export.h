#pragma once
#ifndef MIQT_EXPORT_H
#define MIQT_EXPORT_H

#ifdef _WIN32
	#ifdef MIQT_BUILDING_DLL
		#define MIQT_EXPORT __declspec(dllexport)
	#else
		#define MIQT_EXPORT __declspec(dllimport)
	#endif
	#define MIQT_CALL __cdecl
#else
	#define MIQT_EXPORT __attribute__((visibility("default")))
	#define MIQT_CALL
#endif

#endif // MIQT_EXPORT_H
